package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"path"
	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	eml "gitlab.com/vjsideprojects/relay/internal/platform/integration/email"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type EmailNotification struct {
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
		Requester   string
		Body        string
	}{
		Name:        emNotif.Name,
		AccountName: emNotif.AccountName,
		MagicLink:   emNotif.MagicLink,
		Requester:   emNotif.Requester,
		Body:        emNotif.Body,
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

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	dir := path.Join(path.Dir(basepath), "..")
	log.Println("dir ", dir)

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
