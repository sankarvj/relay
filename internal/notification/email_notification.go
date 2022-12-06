package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"

	"github.com/jmoiron/sqlx"
	eml "gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type EmailNotification struct {
	AccountID   string
	To          []interface{}
	Subject     string
	Body        string
	Name        string
	Requester   string
	AccountName string
	MagicLink   string
}

func (emNotif EmailNotification) Send(ctx context.Context, notifType NotificationType, db *sqlx.DB) error {
	// send email using aws SES if email notification enabled.
	// from: no-reply@workbaseone.com
	// to: <individual user who got assigned> , <updates to the user if already assigned>, <@mention on the notes/conversations>
	templateData := struct {
		Name        string
		AccountName string
		MagicLink   string
		Unsubscribe string
		Requester   string
		Body        string
		Subject     string
	}{
		Name:        emNotif.Name,
		AccountName: emNotif.AccountName,
		MagicLink:   emNotif.MagicLink,
		Unsubscribe: fmt.Sprintf("https://workbaseone.com/v1/unsubscribe?account_id=%s&email=%s", emNotif.AccountID, toMail(emNotif.To)),
		Requester:   emNotif.Requester,
		Body:        emNotif.Body,
		Subject:     emNotif.Subject,
	}

	template := "welcome.html"
	switch notifType {
	case TypeWelcome:
		template = "welcome.html"
	case TypeMemberInvitation:
		template = "invitation.html"
	case TypeVisitorInvitation:
		template = "visitor_invitation.html"
	default:
		template = "update.html"
	}

	//check me before deployment
	// _, b, _, _ := runtime.Caller(0)
	// basepath := filepath.Dir(b)
	// dir := path.Join(path.Dir(basepath), "..")
	// log.Println("dir ", dir)
	localTesting := false
	// for _, toEmail := range emNotif.To {
	// 	if toEmail == "vijayasankarj@gmail.com" {
	// 		localTesting = true
	// 		break
	// 	}
	// }
	if localTesting {
		log.Println("Magic Link: ", templateData.MagicLink)
		return nil
	}

	err := emNotif.ParseTemplate(fmt.Sprintf("templates/%s", template), templateData)
	if err != nil {
		return err
	}
	e := eml.SESMail{Domain: "", ReplyTo: ""}
	_, err = e.SendMail("WorkbaseONE", "WorkbaseONE <no-reply@workbaseone.com>", "", util.ConvertSliceTypeRev(emNotif.To), emNotif.Subject, emNotif.Body)
	return err
}

func (emNotif *EmailNotification) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	emNotif.Body = buf.String()
	return nil
}

func toMail(to []interface{}) string {
	if len(to) > 0 {
		return to[0].(string)
	}
	return ""
}
