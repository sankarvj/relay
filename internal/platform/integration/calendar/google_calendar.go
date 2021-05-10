package calendar

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type Gcalendar struct {
	OAuthFile string
	TokenJson string
}

func getCalendarConfig(oAuthFile string) (*oauth2.Config, error) {
	return integration.GetConfig(oAuthFile, calendar.CalendarScope)
}

func getCalendarService(oAuthFile, tokenJson string) (*calendar.Service, error) {
	config, err := getCalendarConfig(oAuthFile)
	if err != nil {
		return nil, err
	}

	client, err := integration.Client(config, tokenJson)
	if err != nil {
		return nil, err
	}
	return calendar.New(client)
}

func (g *Gcalendar) EventCreate(calendarID string, meeting *integration.Meeting) error {
	srv, err := getCalendarService(g.OAuthFile, g.TokenJson)
	if err != nil {
		return err
	}

	eventAttendees := make([]*calendar.EventAttendee, len(meeting.Attendees))
	for i, attendeeMail := range meeting.Attendees {
		eventAttendees[i] = &calendar.EventAttendee{Email: attendeeMail}
	}

	event := &calendar.Event{
		Summary:     meeting.Summary,
		Description: meeting.Description,
		Start: &calendar.EventDateTime{
			DateTime: meeting.StartTime,
			TimeZone: meeting.TimeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: meeting.EndTime,
			TimeZone: meeting.TimeZone,
		},
		Attendees: eventAttendees,
	}

	event, err = srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return errors.Wrapf(err, "Unable to create event")
	}
	log.Printf("event --- %+v", event)
	meeting.ID = event.Id
	meeting.CalID = event.ICalUID
	meeting.Created = event.Created
	meeting.Updated = event.Updated
	return nil
}

func (g *Gcalendar) Watch(calendarID, channelID string) error {
	srv, err := getCalendarService(g.OAuthFile, g.TokenJson)
	if err != nil {
		return err
	}

	channel := &calendar.Channel{
		Id:      channelID,
		Type:    "web_hook",
		Address: "https://vjrelay.ngrok.io/notifications",
	}

	watchCall := srv.Acl.Watch(calendarID, channel)
	_, err = watchCall.Do()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gcalendar) Sync(calendarID string, syncToken string) (string, error) {
	srv, err := getCalendarService(g.OAuthFile, g.TokenJson)
	if err != nil {
		return syncToken, err
	}

	t := time.Now().Format(time.RFC3339)

	pageToken := ""
	for {
		evl := srv.Events.List(calendarID)
		if syncToken != "" {
			evl.SyncToken(syncToken)
		} else {
			evl.TimeMin(t)
		}
		evl.PageToken(pageToken)
		evl.ShowDeleted(true).SingleEvents(true).MaxResults(10)
		events, err := evl.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
			//delete token and retry
		}

		if len(events.Items) == 0 {
			fmt.Println("No upcoming events found.")
		} else {
			for _, item := range events.Items {
				date := item.Start.DateTime
				if date == "" {
					date = item.Start.Date
				}
				fmt.Printf("%v (%v)\n", item.Summary, date)
			}
		}
		if events.NextPageToken == "" {
			syncToken = events.NextSyncToken
			break
		}
		pageToken = events.NextPageToken
	}

	log.Println("syncToken --> ", syncToken)

	return syncToken, nil
}
