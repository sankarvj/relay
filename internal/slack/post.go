package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func PostMessage(url string, msg string) error {
	postBody, _ := json.Marshal(map[string]string{
		"text": msg,
	})
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(postBody))
	//Handle Error
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
