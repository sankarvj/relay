package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	viewPublishURL = "https://slack.com/api/views.publish"
)

type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type Element struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Block struct {
	Type     string    `json:"type"`
	Text     *Text     `json:"text,omitempty"`
	Elements []Element `json:"elements,omitempty"`
}

type View struct {
	Type   string  `json:"type"`
	Blocks []Block `json:"blocks"`
}

type HomeView struct {
	UserID string `json:"user_id"`
	View   View   `json:"view"`
}

func MakeHomeView(userID string) HomeView {
	return HomeView{
		UserID: userID,
		View: View{
			Type: "home",
			Blocks: []Block{
				{
					Type: "section",
					Text: &Text{
						Type: "mrkdwn",
						Text: "A simple stack of blocks for the simple sample Block Kit Home tab.",
					},
				},
				{
					Type: "actions",
					Text: nil,
					Elements: []Element{
						{
							Type: "button",
							Text: Text{
								Type:  "plain_text",
								Text:  "Action A",
								Emoji: true,
							},
						},
						{
							Type: "button",
							Text: Text{
								Type:  "plain_text",
								Text:  "Action B",
								Emoji: true,
							},
						},
					},
				},
			},
		},
	}
}

func UpdateHomeView(botToken string, hv HomeView) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	postBody, err := json.Marshal(hv)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", viewPublishURL, bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", botToken))
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	var sr SlackViewResponse
	b, _ := ioutil.ReadAll(response.Body)
	err = json.Unmarshal(b, &sr)
	if err != nil {
		return err
	}

	if !sr.Ok {
		return fmt.Errorf("slack server responded with error %s", sr.Error)
	} else {
		log.Println("----------- APP HOME VIEW UPDATED SUCCESSFULLY ------------")
	}
	defer response.Body.Close()
	return nil
}
