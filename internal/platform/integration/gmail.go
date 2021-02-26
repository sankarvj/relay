package integration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func AskGmailAccessURL(ctx context.Context, oAuthFile string) (string, error) {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return "", err
	}
	return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline), nil
}

func GetToken(oAuthFile, code string) (string, error) {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return "", err
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", errors.Wrap(err, "unable to retrieve token from web")
	}
	tokenJson, err := json.Marshal(tok)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal token")
	}

	return string(tokenJson), nil
}

func WatchMessage(oAuthFile, tokenJson, topicName string) (string, error) {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return "", err
	}

	var emailAddress string
	client, err := client(config, tokenJson)
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
	log.Printf("started watching the user %s", emailAddress)
	return emailAddress, nil
}

func History(oAuthFile, tokenJson string, user string, historyID uint64) error {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return err
	}

	client, err := client(config, tokenJson)
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
	log.Printf("rgmsg %+v for historyID %d", rgmsg, historyID)
	for _, history := range rgmsg.History {
		for _, m := range history.Messages {
			log.Printf("m ID ----> %s", m.Id)
			msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
			if err != nil {
				return err
			}
			log.Printf("msg msg %+v", msg)
			raw, err := base64.StdEncoding.DecodeString(msg.Raw)
			if err != nil {
				return errors.Wrap(err, "decoding message payload")
			}
			log.Printf("msg msg msg msg msg msg msg msg  %+v", raw)

			if len(msg.Payload.Parts) > 0 {
				log.Printf("msg payload %+v", msg.Payload)
				for _, part := range msg.Payload.Parts {
					log.Printf("msg part %+v", part)
					if part.MimeType == "text/html" {
						data, err := base64.StdEncoding.DecodeString(part.Body.Data)
						log.Printf("msg data %+v", data)
						log.Printf("msg err %+v", err)
						if err != nil {
							return errors.Wrap(err, "decoding message payload")
						}
						html := string(data)
						fmt.Println("JTML1-->", html)
					} else if part.MimeType == "text/plain" {
						data, err := base64.StdEncoding.DecodeString(part.Body.Data)
						log.Printf("msg data %+v", data)
						log.Printf("msg err %+v", err)
						if err != nil {
							return errors.Wrap(err, "decoding message payload")
						}
						html := string(data)
						fmt.Println("JTML1-->", html)
					}
				}
			} else {
				log.Printf("msg part %+v", msg.Payload.Body.Data)
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

func sendViaGmail(oAuthFile, tokenJson string, user string, fromName, fromEmail string, toName string, toEmail []string, subject string, body string) (*string, error) {
	config, err := getConfig(oAuthFile)
	if err != nil {
		return nil, err
	}

	client, err := client(config, tokenJson)
	if err != nil {
		return nil, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}

	gmsg := msg(fromName, fromEmail, toName, toEmail[0], subject, body) //TODO how to send multiple to address
	rgmsg, err := srv.Users.Messages.Send(user, &gmsg).Do()
	if err != nil {
		return nil, err
	}
	log.Printf("rgmsg %+v", rgmsg.ThreadId)
	return &rgmsg.ThreadId, nil
}

func getConfig(oAuthFile string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(oAuthFile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read client secret file")
	}
	return google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailSendScope)
}

func client(config *oauth2.Config, tokenJson string) (*http.Client, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenJson), &token); err != nil {
		return nil, err
	}
	return config.Client(context.Background(), &token), nil
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
