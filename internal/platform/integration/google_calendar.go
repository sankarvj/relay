package integration

import (
	"fmt"
	"log"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

func getCalendarConfig(oAuthFile string) (*oauth2.Config, error) {
	return getConfig(oAuthFile, calendar.CalendarScope)
}

func CreateEvent(oAuthFile string, tokenJson string) error {
	config, err := getCalendarConfig(oAuthFile)
	if err != nil {
		return err
	}

	client, err := client(config, tokenJson)
	if err != nil {
		return err
	}

	srv, err := calendar.New(client)
	if err != nil {
		return err
	}

	event := &calendar.Event{
		Summary:     "Google I/O 2021",
		Location:    "800 Howard St., San Francisco, CA 94103",
		Description: "A chance to hear more about Google's developer products.",
		Start: &calendar.EventDateTime{
			DateTime: "2021-05-28T09:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: "2021-05-28T17:00:00-07:00",
			TimeZone: "America/Los_Angeles",
		},
		Recurrence: []string{"RRULE:FREQ=DAILY;COUNT=2"},
		Attendees: []*calendar.EventAttendee{
			{Email: "vijayasankarmobile@gmail.com"},
			{Email: "sbrin@example.com"},
		},
	}

	calendarId := "primary"
	event, err = srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.HtmlLink)

	return nil
}
