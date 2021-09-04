package email

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

type Gmail struct {
	OAuthFile string
	TokenJson string
}

func (g Gmail) Watch(topicName string) (string, error) {
	config, err := getGmailConfig(g.OAuthFile)
	if err != nil {
		return "", err
	}

	var emailAddress string
	client, err := integration.Client(config, g.TokenJson)
	if err != nil {
		return emailAddress, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return emailAddress, err
	}

	user := "me"
	profileCall := srv.Users.GetProfile(user)
	profile, err := profileCall.Do()
	if err != nil {
		return emailAddress, err
	}
	emailAddress = profile.EmailAddress

	watchCall := srv.Users.Watch(emailAddress, &gmail.WatchRequest{
		TopicName: topicName,
	})
	_, err = watchCall.Do()
	if err != nil {
		return emailAddress, err
	}
	log.Printf("internal.platform.integration.email.gmail started watching the user %s\n", emailAddress)
	return emailAddress, nil
}

func Message(oAuthFile, tokenJson string, user string, messageID string) error {
	config, err := getGmailConfig(oAuthFile)
	if err != nil {
		return err
	}

	client, err := integration.Client(config, tokenJson)
	if err != nil {
		return err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return err
	}

	rgmsg, err := srv.Users.Threads.Get("me", messageID).Do()
	if err != nil {
		log.Println("rgmsg err ", err)
		return err
	}
	log.Printf("rgmsg %+v ", rgmsg)

	return nil
}

func History(oAuthFile, tokenJson string, user string, historyID uint64) error {
	config, err := getGmailConfig(oAuthFile)
	if err != nil {
		return err
	}

	client, err := integration.Client(config, tokenJson)
	if err != nil {
		return err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return err
	}

	rgmsg, err := srv.Users.History.List(user).StartHistoryId(historyID).Do()
	if err != nil {
		return err
	}
	for _, history := range rgmsg.History {
		for _, m := range history.Messages {
			msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
			if err != nil {
				return err
			}

			log.Printf("internal.platform.integration.email.gmail message  %+v\n", msg)

			if len(msg.Payload.Parts) > 0 {

				log.Println("msg snippet ", msg.Snippet)
				log.Println("msg threadID ", msg.ThreadId)

				for _, part := range msg.Payload.Parts {
					if part.MimeType == "text/html" {
						data, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							return errors.Wrap(err, "decoding message payload")
						}
						html := string(data)
						fmt.Println("JTML1-->", html)
					} else if part.MimeType == "text/plain" {
						data, err := base64.StdEncoding.DecodeString(part.Body.Data)
						if err != nil {
							return errors.Wrap(err, "decoding message payload")
						}
						html := string(data)
						fmt.Println("JTML1-->", html)
					}
				}
			} else {
				data, err := base64.StdEncoding.DecodeString(msg.Payload.Body.Data)
				if err != nil {
					return errors.Wrap(err, "decoding message payload")
				}
				html := string(data)
				fmt.Println("JTML2-->", html)
			}

		}
	}

	return nil
}

func (g Gmail) SendMail(fromName, fromEmail string, toName string, toEmail []string, subject string, body string) (*string, error) {
	config, err := getGmailConfig(g.OAuthFile)
	if err != nil {
		return nil, err
	}

	client, err := integration.Client(config, g.TokenJson)
	if err != nil {
		return nil, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}

	gmsg := msg(fromName, fromEmail, toName, toEmail[0], subject, body) //TODO how to send multiple to address
	rgmsg, err := srv.Users.Messages.Send("me", &gmsg).Do()
	if err != nil {
		return nil, err
	}
	log.Printf("rgmsg %+v", rgmsg.ThreadId)
	return &rgmsg.ThreadId, nil
}

func getGmailConfig(oAuthFile string) (*oauth2.Config, error) {
	return integration.GetConfig(oAuthFile, integration.GmailScopes...)
}

func msg(fromName, fromEmail string, toName, toEmail string, subject string, body string) gmail.Message {
	from := mail.Address{Name: fromName, Address: fromEmail}
	to := mail.Address{Name: toName, Address: toEmail}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	var msg string
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	fmt.Printf("header --> %+v", header)
	fmt.Println("msg --> ", msg)

	return gmail.Message{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(msg)),
	}
}

type Code struct {
	Code string `json:"code"`
}
