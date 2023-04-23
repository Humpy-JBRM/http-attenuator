package cmd

import (
	"fmt"
	"http-attenuator/data"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var alexa1mCmd = &cobra.Command{
	Use:   "alexa1m",
	Short: "Runs the alexa1m, which just makes requests to sites in the Alexa1m to produce nice varz for grafana",
	Run:   RunAlexa1m,
}

var hsakAddress string
var hertz int

func init() {
	brokerCmd.PersistentFlags().StringVarP(&hsakAddress, "hsak", "h", "localhost:8888", "Address of HSAK gateway (default is localhost:8888)")
	brokerCmd.PersistentFlags().IntVarP(&hertz, "hz", "z", 2, "Site request frequency (default is 2Hz)")
}

func RunAlexa1m(cmd *cobra.Command, args []string) {
	alexa1mBytes, err := os.ReadFile("cmd/alexa1m_sites.txt")
	if err != nil {
		log.Fatal(err)
	}
	alexa1m := make([]string, 0)
	count := 0
	for _, site := range strings.Split(string(alexa1mBytes), "\n") {
		trimmed := strings.TrimSpace(site)
		if trimmed != "" {
			count++
			if count > 10 {
				break
			}
			alexa1m = append(alexa1m, site)
		}
	}
	rand.Seed(time.Now().UnixMilli())
	requestChan := make(chan string, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		go worker(requestChan, &wg)
	}
	for {
		requestChan <- alexa1m[rand.Intn(len(alexa1m))]
	}
}

// This is a contrived bunch of data, used only for the purposes
// of building out the grafana dashboard and the various billing
// mechanisms
//
// TODO(john): this must be a proper customer database
//
// See: middleware/billing.go
var apiKeys = []string{
	"",         // "(unknown)",
	"69696969", // "Ana de Armas",
	"88888888", // "Alizee",
	"666",      // "GCHQ",
}

func worker(requestChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	sleepTimeMillis := int64(1000) / int64(hertz)
	for {
		time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
		site := <-requestChan

		// Choose a API key at random
		apiKey := apiKeys[rand.Intn(len(apiKeys))]

		// ~50% of the time, we want to use the gateway.
		//
		// The rest of the time we want to use the service broker.
		//
		// This is mainly to test that our 'weighted' rule for selecting
		// backends is doing what we expect it to do
		var targetUrl string
		switch rand.Int()%2 == 0 {
		case true:
			// Even number
			// Gateway request
			// Do the request
			siteUrl := "http://" + site
			targetUrl = fmt.Sprintf("http://%s/api/v1/gateway/%s", hsakAddress, siteUrl)

		default:
			// Odd number
			// Use the service broker
			targetUrl = fmt.Sprintf("http://%s/api/v1/broker/alexa1m", hsakAddress)
		}

		u, _ := url.Parse(targetUrl)
		client := &http.Client{}
		req := http.Request{
			Method: "GET",
			URL:    u,
			Header: http.Header{},
		}
		req.Header.Add(data.HEADER_X_MIGALOO_API_KEY, apiKey)
		log.Println(targetUrl)
		_, err := client.Do(&req)
		if err != nil {
			log.Printf("%s: %s", targetUrl, err.Error())
		}
	}
}
