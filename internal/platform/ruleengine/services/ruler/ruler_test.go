package ruler

import (
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	sampleInput := `{{e1.appinfo.version}} eq {{e2.version}} <e3.status=e2.version>`
	//sampleInput = `{{a6036fe2-0e77-4fab-a798-a39fcf99815c.build.appinfo.version}} eq {{8ac6147e-ad53-4379-8503-806c01500b9b.latest.version}} <8ac6147e-ad53-4379-8503-806c01500b9b.latest.status=up>`

	action := make(chan string)
	go Run(sampleInput, getResponseMap, action)

	for {
		act, ok := <-action
		if !ok {
			fmt.Println("Channel Close")
			break
		}
		fmt.Println("Action To Be Taken ", act)
	}

}

func getResponseMap(key string) map[string]interface{} {
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
