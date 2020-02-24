package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//APIParams makes the net call with the parameters it has
type APIParams struct {
	Method  string
	Host    string
	Path    string
	Headers map[string]string
	ReqBody map[string]interface{}
}

//MakeHTTPRequest makes api calls to other servers
func (ap APIParams) MakeHTTPRequest(response *map[string]interface{}) error {
	jsonbody, err := json.Marshal(ap.ReqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(ap.Method, ap.Host+ap.Path, bytes.NewBuffer(jsonbody))
	if err != nil {
		return err
	}
	for key, val := range ap.Headers {
		req.Header.Add(key, val)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("STATUS ----->", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &response)
}
