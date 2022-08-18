package tracker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Track struct {
	project string
	channel string
	icon    string
}

func EmailChan() Track {
	return Track{
		project: "relay",
		channel: "email",
		icon:    "📬",
	}
}

func ErrorChan() Track {
	return Track{
		project: "relay",
		channel: "error",
		icon:    "🐞",
	}
}

func TestChan() Track {
	return Track{
		project: "relay",
		channel: "test",
		icon:    "🦄",
	}
}

func (t Track) MailIcon() Track {
	t.icon = "📬"
	return t
}

func (t Track) Log(event, desc string) {
	notify := "true"
	payload := strings.NewReader(`{
		"project": "` + t.project + `", 
		"channel": "` + t.channel + `",
		"event": "` + event + `",
		"description": "` + desc + `",
		"icon": "` + t.icon + `",
		"notify": "` + notify + `"}`)
	post(payload)
}

func post(payload *strings.Reader) {
	url := "https://api.logsnag.com/v1/log"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println("LogSnag Err", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer 7bab6d5c2729a1a7f88c13ec27d58ef6")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("LogSnag Err", err)
		return
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("LogSnag Err", err)
		return
	}
	//fmt.Println("Tracker body", string(body))
}
