package email

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"

	"gitlab.com/vjsideprojects/relay/internal/platform/integration"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

type Gmail struct {
	OAuthFile string
	TokenJson string
	ReplyTo   string
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
	//calling this as a safer side because calling two watch is not allowed
	srv.Users.Stop(emailAddress).Do()
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

func (g Gmail) Stop(emailAddress string) error {
	log.Println("Stop Called!")
	config, err := getGmailConfig(g.OAuthFile)
	if err != nil {
		return err
	}

	client, err := integration.Client(config, g.TokenJson)
	if err != nil {
		return err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return err
	}

	srv.Users.Stop(emailAddress)

	return nil
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
		return err
	}
	log.Printf("rgmsg %+v ", rgmsg)

	return nil
}

func History(oAuthFile, tokenJson string, user string, historyID uint64) (uint64, error) {
	config, err := getGmailConfig(oAuthFile)
	if err != nil {
		return 0, err
	}

	client, err := integration.Client(config, tokenJson)
	if err != nil {
		return 0, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return 0, err
	}

	rgmsg, err := srv.Users.History.List(user).HistoryTypes("messageAdded").StartHistoryId(historyID).Do()
	if err != nil {
		return 0, err
	}

	log.Println("historyID ", historyID)
	if len(rgmsg.History) == 0 {
		log.Println("Why the heck the history of messages are empty!!!!!!!!!!!")
	}

	for _, history := range rgmsg.History {
		for _, m := range history.Messages {
			msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
			if err != nil {
				return 0, err
			}

			log.Println("msg snippet ", msg.Snippet)
			log.Println("msg threadID ", msg.ThreadId)

			for i := 0; i < len(msg.Payload.Headers); i++ {
				headerVal := msg.Payload.Headers[i]
				log.Println("msg Header Name ", headerVal.Name)
				log.Println("msg Header Value ", headerVal.Value)
			}

		}
	}

	return rgmsg.HistoryId, nil
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
