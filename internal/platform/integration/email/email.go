package email

type Email interface {
	SendMail(fromName, fromEmail string, toName string, toEmail []string, subject string, body string) (*string, error)
	Watch(topicName string) (string, error)
}
