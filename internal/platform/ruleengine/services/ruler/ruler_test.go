package ruler

import (
	"log"
	"testing"
)

func TestRun(t *testing.T) {
	sampleInput := `{{e1.appinfo.version}} eq {{e2.version}} && {{e1.appinfo.version1}} eq {{e2.version}} <e3.status=e2.version>`

	signalsChan := make(chan Work)
	go Run(sampleInput, signalsChan)
	//signalsChan wait to receive work and action triggers until the run completes
	for work := range signalsChan {
		if work.Resp != nil { //is it a right way to differentiate the expression work and action expression?
			work.Resp <- getResponseMap(work.Expression)
		} else {
			log.Println("trigger>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", work.Expression)
		}
	}
	log.Println("signals channel closed!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func getResponseMap(exp string) map[string]interface{} {
	key := FetchEntityID(exp)
	if key == "e1" {
		return map[string]interface{}{
			"e1": map[string]interface{}{
				"artifact": 1,
				"appinfo": map[string]interface{}{
					"version": 2,
				},
			},
		}
	} else {
		return map[string]interface{}{
			"e2": map[string]interface{}{
				"version": 2,
			},
		}
	}
}
