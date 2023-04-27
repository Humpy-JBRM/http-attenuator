package server

import (
	"fmt"
	"http-attenuator/data"
	config "http-attenuator/facade/config"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
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
		err := fmt.Errorf("%s.Handle(): could not get handler", p.name)
		log.Println(err)
		c.AbortWithError(
			http.StatusInternalServerError,
			err,
		)
		return
	}

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

func NewPathologyRegistryFromConfig(pathologyRoot map[string]interface{}) error {
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

func parseHttpTimeouts(pathology *FailureModeImpl) error {
	valuesRoot := data.CONF_PATHOLOGY + "." + pathology.name + ".httpcode"

	// The values is a map[string]interface{}
	// httpcode:
	// 	"404":
	// 	weight: 1
	//   "429":
	// 	weight: 5
	// 	# The headers to return when we encounter this code
	// 	headers:
	// 	  X-Backoff-Millis: 60000
	// 	  X-Retry-After: ${now} + 60s
	httpCodes, err := config.Config().GetAllValues(valuesRoot)
	if err != nil {
		return err
	}
	if len(httpCodes) == 0 {
		return fmt.Errorf("parseHttpTimeouts(%s): '%s' has no values", pathology.name, valuesRoot)
	}

	// for codeAsString := range httpCodes {
	// 	codeAsInt, err := strconv.Atoi(codeAsString)
	// 	if err != nil {
	// 		return fmt.Errorf("parseHttpTimeouts(%s): '%s' is not a numeric httpcode", pathology.name, codeAsString)
	// 	}

	// 	valuesRoot := configName + "." + codeAsString
	// 	values, err := config.Config().GetAllValues(valuesRoot)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// configName = valuesRoot + ".timeout_millis"
	// 	// pathology.timeoutMillis, err = config.Config().GetInt(configName)
	// 	// if err != nil {
	// 	// 	return pathologies, err
	// 	// }
	// }

	return nil
}
