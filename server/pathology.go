package server

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var pathologyRequests = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_requests",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_errors",
		Help:      "The requests handled by the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyLatency = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_latency",
		Help:      "The latency of the various pathologies, keyed by name, handler and method",
	},
	[]string{"pathology", "handler", "method"},
)
var pathologyResponses = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "migaloo",
		Name:      "pathology_responses",
		Help:      "The responses from the various pathologies, keyed by name, method and status code",
	},
	[]string{"pathology", "handler", "method", "code"},
)

type FailureMode interface {
	data.HasCDF
	Handler
}

type FailureModeImpl struct {
	name    string
	weight  int64
	cdf     float64
	handler Handler
}

func (f *FailureModeImpl) GetName() string {
	return f.name
}

func (f *FailureModeImpl) Handle(c *gin.Context) {
	f.handler.Handle(c)
}

// make FailureModeImpl conform to the HasCDF interface
// so we can use our generic ChooseFromCDF() function
func (f *FailureModeImpl) CDF() float64 {
	return f.cdf
}

type FailureModeDistribution struct {
	failureModes []data.HasCDF
}

func (pd *FailureModeDistribution) ChooseFailureMode() FailureMode {
	choice := data.ChooseFromCDF(rand.Float64(), pd.failureModes)
	if choice == nil {
		return nil
	}

	return choice.(FailureMode)
}

type Pathology interface {
	Handler
	ChooseFailureMode() FailureMode
}

type PathologyImpl struct {
	name         string
	failureModes FailureModeDistribution
}

func (p *PathologyImpl) GetName() string {
	return p.name
}

func (p *PathologyImpl) ChooseFailureMode() FailureMode {
	return p.failureModes.ChooseFailureMode()
}

func (p *PathologyImpl) Handle(c *gin.Context) {
	failureMode := p.ChooseFailureMode()
	if failureMode == nil {
		pathologyErrors.WithLabelValues(p.name, failureMode.GetName(), c.Request.Method).Inc()
		err := fmt.Errorf("%s.Handle(): could not get handler", p.name)
		log.Println(err)
		c.AbortWithError(
			http.StatusInternalServerError,
			err,
		)
		return
	}

	c.Set("pathology", p.name)
	failureMode.Handle(c)
}

type PathologyRegistry interface {
	GetPathology(name string) Pathology
}

var pathologyRegistry PathologyRegistry

func GetPathologyRegistry() PathologyRegistry {
	return pathologyRegistry
}

type PathologyRegistryImpl struct {
	pathologiesByName map[string]Pathology
}

func (r *PathologyRegistryImpl) GetPathology(name string) Pathology {
	return r.pathologiesByName[strings.ToLower(name)]
}

func NewPathologyRegistryFromConfig() error {
	pathologyRegistryImpl := &PathologyRegistryImpl{
		pathologiesByName: make(map[string]Pathology),
	}
	// Get everything under config.pathology
	pathologyRoot, err := config.Config().GetAllValues(data.CONF_PATHOLOGY)
	if err != nil {
		return err
	}

	for pathologyName, _ := range pathologyRoot {
		pathology := &PathologyImpl{
			name: pathologyName,
		}

		// The values is a map[string]interface{}
		// failure_weights:
		err = parseFailureWeights(pathology)
		if err != nil {
			return err
		}

		for i := 0; i < len(pathology.failureModes.failureModes); i++ {
			switch pathology.failureModes.failureModes[i].(*FailureModeImpl).name {
			case "httpcode":
				err = parseHttpCode(pathology)
				if err != nil {
					return err
				}

			case "timeout":

			default:
				return fmt.Errorf("NewPathologyRegistryFromConfig(): unrecognised failure mode: '%s'", pathology.failureModes.failureModes[i].(*FailureModeImpl).name)
			}
		}

		// valuesRoot := data.CONF_PATHOLOGY + "." + pathologyName
		// // httpcode:
		// configName := valuesRoot + ".httpcode"
		// httpCodes, err := config.Config().GetAllValues(configName)
		// if err != nil {
		// 	return pathologies, err
		// }
		// if len(httpCodes) == 0 {
		// 	return pathologies, fmt.Errorf("NewPathologyFromConfig(): '%s' has no values", configName)
		// }

		// // timeout_millis: 10000
		// configName = valuesRoot + ".timeout_millis"
		// timeoutMillis, err := config.Config().GetInt(configName)
		// if err != nil {
		// 	return pathologies, err
		// }

		// Stash this pathology
		pathologyRegistryImpl.pathologiesByName[pathologyName] = pathology
	}

	pathologyRegistry = pathologyRegistryImpl
	return nil
}

// pathology:
// simple:
//
//	  # The failure pathology
//	  #
//	  #   pathology: weight
//	  failure_weights:
//		httpcode: 90
//		timeout: 10
func parseFailureWeights(pathology *PathologyImpl) error {
	valuesRoot := data.CONF_PATHOLOGY + "." + pathology.name + ".failure_weights"
	failureWeights, err := config.Config().GetAllValues(valuesRoot)
	if err != nil {
		return err
	}
	if len(failureWeights) == 0 {
		return fmt.Errorf("parseFailureWeights(%s): '%s' has no values", pathology.name, valuesRoot)
	}

	var totalWeight int64
	for failureModeName := range failureWeights {
		weight, err := config.Config().GetInt(valuesRoot + "." + failureModeName)
		if err != nil {
			return err
		}
		failureMode := &FailureModeImpl{
			name:   failureModeName,
			weight: weight,
		}
		pathology.failureModes.failureModes = append(pathology.failureModes.failureModes, failureMode)
		totalWeight += weight
	}

	// Calculate the cdf for the failure modes
	var totalProbability float64
	for i := 0; i < len(pathology.failureModes.failureModes); i++ {
		probability := float64(float64(pathology.failureModes.failureModes[i].(*FailureModeImpl).weight) / float64(totalWeight))
		totalProbability += probability
		pathology.failureModes.failureModes[i].(*FailureModeImpl).cdf = totalProbability
	}

	sort.Slice(pathology.failureModes.failureModes, func(i, j int) bool {
		return pathology.failureModes.failureModes[i].(*FailureModeImpl).name < pathology.failureModes.failureModes[j].(*FailureModeImpl).name
	})
	return nil
}

// parseHttpCode reads in and parses the
// config:
//
//	pathology:
//	  # Pathologies have names.
//	  #
//	  # This allows us to easily create pathologies which have specific
//	  # behaviour and then just refer to them by name in a particular
//	  # server config.
//	  #
//	  # This ability will become even more imortant / useful as we
//	  # extend the config API to allow backend servers to be created
//	  # and configured programatically
//	  simple:
//	    # The failure pathology
//	    #
//	    #   pathology: weight
//	    failure_weights:
//	      httpcode: 90
//	      timeout: 10
//	    # The http code pathology
//	    httpcode:
//	      # The HTTP codes to return, and the weight for each return code.
//	      # The weights do not need to add up to 100, I just made them add
//	      # up to 100 here so its easy to grok the % of time that code is returned
//	      "200":
//	        duration: normal(1, 0.2)
//	        weight: 80
//	        headers:
//	          Content-type: application/json
//	        body: {"success": true}
//	      "401":
//	        weight: 5
//	      "404":
//	        weight: 1
//	      "429":
//	        weight: 5
//	        # The headers to return when we encounter this code
//	        headers:
//	          X-Backoff-Millis: 60000
//	          X-Retry-After: now() + 60s
func parseHttpCode(pathology *PathologyImpl) error {
	valuesRoot := data.CONF_PATHOLOGY + "." + pathology.name + ".httpcode"
	httpCodeMap, err := config.Config().GetAllValues(valuesRoot)
	if err != nil {
		return err
	}
	if len(httpCodeMap) == 0 {
		return fmt.Errorf("parseHttpCode(%s): '%s' has no values", pathology.name, valuesRoot)
	}

	var totalWeight int64
	normalDurationRegex := regexp.MustCompile(`normal\(([0-9]+.[0-9]+), ([0-9]+.[0-9]+)\)`)
	cdf := make([]data.HasCDF, 0)
	for httpCode := range httpCodeMap {
		builder := NewHttpCodeCdfBuilder()
		numericCode, err := strconv.Atoi(httpCode)
		if err != nil {
			return fmt.Errorf("%s: code '%s': %s", valuesRoot+"."+httpCode, httpCode, err.Error())
		}
		builder.Code(numericCode)

		// Is there any duration specified?
		//
		// TODO(john): a proper grammar, NOT brittle and janky regex
		duration, err := config.Config().GetString(valuesRoot + "." + httpCode + ".duration")
		if err != nil {
			return err
		}
		var mean float64
		var stddev float64
		if duration != "" {
			if !normalDurationRegex.MatchString(duration) {
				return fmt.Errorf("%s: duration '%s' does not match %s", valuesRoot+"."+httpCode, duration, normalDurationRegex)
			}
			matches := normalDurationRegex.FindStringSubmatch(duration)
			if len(matches) != 3 {
				return fmt.Errorf("%s: duration '%s' does not match %s", valuesRoot+"."+httpCode, duration, normalDurationRegex)
			}
			mean, err = strconv.ParseFloat(matches[1], 64)
			if err != nil {
				return fmt.Errorf("%s: mean '%s': %s", valuesRoot+"."+httpCode, matches[1], err.Error())
			}
			if mean <= 0 {
				return fmt.Errorf("%s: mean '%s': must be > 0", valuesRoot+"."+httpCode, matches[1])
			}
			stddev, err = strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return fmt.Errorf("%s: stddev '%s': %s", valuesRoot+"."+httpCode, matches[2], err.Error())
			}
			if stddev <= 0 {
				return fmt.Errorf("%s: stddev '%s': must be > 0", valuesRoot+"."+httpCode, matches[1])
			}
			builder.DurationMean(mean)
			builder.DurationStddev(stddev)
		}

		weight, err := config.Config().GetInt(valuesRoot + "." + httpCode + ".weight")
		if err != nil {
			return err
		}
		builder.Weight(int(weight))
		totalWeight += weight

		// Any headers?
		headers, err := config.Config().GetValue(valuesRoot + "." + httpCode + ".headers")
		if err != nil {
			return err
		}
		if headers != nil {
			for k, v := range headers.(map[string]interface{}) {
				builder.AddHeader(k, fmt.Sprint(v))
			}
		}

		// Any response body?
		responseBody, err := config.Config().GetValue(valuesRoot + "." + httpCode + ".body")
		if err != nil {
			return err
		}
		if responseBody != nil {
			builder.Body([]byte(fmt.Sprint(responseBody)))
		}

		httpCodeCdf := builder.Build()
		cdf = append(cdf, httpCodeCdf)
	}

	// Backpatch the CDF
	var totalProbability float64
	for i := 0; i < len(cdf); i++ {
		totalProbability += float64(float64(cdf[i].(*HttpCodeCdf).weight) / float64(totalWeight))
		cdf[i].(*HttpCodeCdf).cdf = totalProbability
	}
	sort.Slice(cdf, func(i, j int) bool {
		return cdf[i].(*HttpCodeCdf).cdf < cdf[j].(*HttpCodeCdf).cdf
	})

	// create the httpCode handler
	httpCodeHandler := &HttpCodeHandler{
		BaseHandler: BaseHandler{
			name: "httpcode",
		},
		cdf: cdf,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// associate the handler with its failure mode
	for i := 0; i < len(pathology.failureModes.failureModes); i++ {
		if pathology.failureModes.failureModes[i].(*FailureModeImpl).name == "httpcode" {
			pathology.failureModes.failureModes[i].(*FailureModeImpl).handler = httpCodeHandler
			break
		}
	}

	return nil
}
