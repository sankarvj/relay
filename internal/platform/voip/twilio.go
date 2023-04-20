package voip

import (
	"fmt"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/twilio/twilio-go/twiml"
)

func MakeCall(twilioSID, twilioToken, accountID string, accountToken string, entityID, itemID, tophone string) error {

	// Find your Account SID and Auth Token at twilio.com/console
	// and set the environment variables. See http://twil.io/secure
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: twilioSID,
		Password: twilioToken,
	})
	params := &api.CreateCallParams{}

	webhook := fmt.Sprintf("https://a9ef-2601-646-9901-6720-9042-8f36-a3cc-653a.ngrok-free.app/v1/accounts/%s/twilio/%s/entities/%s/items/%s", accountID, accountToken, entityID, itemID)

	sayIncident := &twiml.VoiceSay{
		Message:  "You got an incident",
		Voice:    "woman",
		Language: "en-US",
	}

	pause := &twiml.VoicePause{
		Length: "1",
	}

	sayBody := &twiml.VoiceSay{
		Message:  "To acknoweldge this incident. Please press number 1",
		Voice:    "woman",
		Language: "en-US",
	}

	gather := &twiml.VoiceGather{
		Action:              webhook,
		ActionOnEmptyResult: "true",
		Input:               "dtmf",
		NumDigits:           "1",
		Timeout:             "4",
	}

	verbList := []twiml.Element{sayIncident, pause, sayBody, gather}
	twimlResult, err := twiml.Voice(verbList)
	if err != nil {
		return err
	}

	fmt.Printf("twimlResult :::: %+v", twimlResult)

	params.SetTwiml(twimlResult)
	params.SetTo(tophone)
	params.SetFrom("+18556761926")
	params.SetStatusCallback(webhook)
	params.SetStatusCallbackEvent([]string{"ringing"})

	resp, err := client.Api.CreateCall(params)
	if err != nil {
		fmt.Println(err.Error())
		return err
	} else {
		if resp.Sid != nil {
			fmt.Printf("Resp :::: %+v", *resp)
		} else {
			fmt.Println(resp.Sid)
		}
		return nil
	}
}
