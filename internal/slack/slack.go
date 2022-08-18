package slack

import (
	"log"
)

const (
	EventTypeHomeOpened string = "app_home_opened"
	EventTypeMention    string = "app_mention"
	EventTypeMessage    string = "message"
)

func (s Slack) Call() error {
	switch s.Event.Type {
	case EventTypeHomeOpened:
		return s.appHomeOpened()
	case EventTypeMention:
		return s.appMention()
	case EventTypeMessage:
		return s.appMessage()
	default:
		return s.appDefault()
	}
}

func (s Slack) appHomeOpened() error {
	log.Println("----------- APP HOME OPENED ------------")
	log.Printf("the slack payload::  %+v \n", s.Payload)
	return UpdateHomeView(s.BotToken, MakeHomeView(s.Event.User))
}

func (s Slack) appMention() error {
	log.Println("----------- APP MENTION ------------")
	log.Printf("the slack payload::  %+v \n", s.Payload)
	return nil
}

func (s Slack) appMessage() error {
	log.Println("----------- APP MESSAGE ------------")
	log.Printf("the slack payload::  %+v \n", s.Payload)
	return nil
}

func (s Slack) appDefault() error {
	log.Println("--------- NOT IMPLEMENTED ----------")
	log.Printf("the slack payload::  %+v \n", s.Payload)
	return nil
}
