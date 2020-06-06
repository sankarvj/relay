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
		switch work.Type {
		case Worker:
			work.Resp <- getResponseMap(work.Expression)
		case Executor:
			log.Println("trigger>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", work.Expression)
		case Content:
			log.Println("content>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", work.Expression)
		}
	}
	log.Println("signals channel closed!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func TestRunSimpleBody(t *testing.T) {
	sampleInput := `Hello matty {{e1.appinfo.version}}. How are you?`

	signalsChan := make(chan Work)
	go Run(sampleInput, signalsChan)
	//signalsChan wait to receive work and action triggers until the run completes
	for work := range signalsChan {
		switch work.Type {
		case Worker:
			work.Resp <- getResponseMap(work.Expression)
		case Executor:
			log.Println("trigger>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", work.Expression)
		case Content:
			log.Println("content>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", work.Expression)
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
