package data

import (
	config "http-attenuator/facade/config"
	"os"
	"testing"
)

func TestBrokerFromConfig(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../test_resources/broker/config.yml")
	config.Config()

	appConfig, err := LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		t.Fatal(err)
	}

	broker := appConfig.Config.Broker
	expectedListen := "0.0.0.0:8888"
	actualListen := broker.Listen
	if expectedListen != actualListen {
		t.Errorf("Expected listen='%s', but got '%s'", expectedListen, actualListen)
	}
	expectedLen := 3
	actualLen := len(broker.upstream)
	if expectedLen != actualLen {
		t.Errorf("Expected %d upstream services, but got %d", expectedLen, actualLen)
	}

	// TODO(john): the ChooseFromCDF() seems to be off ...
	// Fix it.
	// alexa1mUpstream := broker.GetUpstream("alexa1m")
	// actualChosenBackend := make(map[string]int)
	// expectedChosenBackend := map[string]float64{
	// 	"amazon.com":    0.08,
	// 	"baidu.com":     0.08,
	// 	"bilibili.com":  0.08,
	// 	"facebook.com":  0.19,
	// 	"google.com":    0.1,
	// 	"qq.com":        0.08,
	// 	"twitter.com":   0.08,
	// 	"wikipedia.org": 0.08,
	// 	"youtube.com":   0.14,
	// 	"zhihu.com":     0.08,
	// }
	// iterations := 100000
	// for i := 0; i < iterations; i++ {
	// 	chosenBackend := alexa1mUpstream.ChooseBackend()
	// 	if chosenBackend == nil {
	// 		t.Errorf("%s.ChooseBackend() returned nil", alexa1mUpstream.GetName())
	// 	}
	// 	actualChosenBackend[chosenBackend.GetName()]++
	// }
	// for chosenName := range expectedChosenBackend {
	// 	if !util.AlmostEqual(expectedChosenBackend[chosenName], float64(actualChosenBackend[chosenName])/float64(iterations)) {
	// 		t.Errorf("Expected %s chosen ~%.2f, but it was %0.2f", chosenName, expectedChosenBackend[chosenName], float64(actualChosenBackend[chosenName])/float64(iterations))
	// 	}
	// }
}
