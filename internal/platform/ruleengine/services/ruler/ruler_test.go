package ruler

import (
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	sampleInput := `{{e1.appinfo.version}} eq {{e2.version}} || {{e1.appinfo.version1}} eq {{e2.version}} <e3.status=e2.version>`

	triggerChan := make(chan string)
	workChan := make(chan Work)
	go Run(sampleInput, workChan, triggerChan)
	go startWorker(workChan)
	for {
		act, ok := <-triggerChan
		if !ok {
			fmt.Println("Channel Close 1")
			break
		}
		fmt.Println("Action To Be Taken ", act)
	}

}

func startWorker(w chan Work) {
	for {
		work, ok := <-w
		if !ok {
			fmt.Println("Channel Close 2")
			break
		}
		work.Resp <- getResponseMap(work.Expression)
	}
}

func getResponseMap(exp string) map[string]interface{} {
	key := FetchRootKey(exp)
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
