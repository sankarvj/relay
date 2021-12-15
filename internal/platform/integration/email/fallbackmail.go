package email

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	AccessKey = ""
	SecretKey = ""
	// Replace sender@example.com with your "From" address.
	// This address must be verified with Amazon SES.
	Sender = "contact@wayplot.com"

	// Replace recipient@example.com with a "To" address. If your account
	// is still in the sandbox, this address must be verified.
	Recipient = "vijayasankarmail@gmail.com"

	// Specify a configuration set. To use a configuration
	// set, comment the next line and line 92.
	//ConfigurationSet = "ConfigSet"

	// The subject line for the email.
	Subject = "Amazon SES Test (AWS SDK for Go)"

	// The HTML body for the email.
	HtmlBody = "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"

	//The email body for recipients with non-HTML email clients.
	TextBody = "This email was sent with Amazon SES using the AWS SDK for Go."

	// The character encoding for the email.
	CharSet = "UTF-8"
)

type FallbackMail struct {
	Domain string
}

type MailBody struct {
	NotificationType string `json:"notificationType"`
	Mail             struct {
		Timestamp        time.Time `json:"timestamp"`
		Source           string    `json:"source"`
		MessageID        string    `json:"messageId"`
		Destination      []string  `json:"destination"`
		HeadersTruncated bool      `json:"headersTruncated"`
		Headers          []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
		CommonHeaders struct {
			ReturnPath string   `json:"returnPath"`
			From       []string `json:"from"`
			Date       string   `json:"date"`
			To         []string `json:"to"`
			MessageID  string   `json:"messageId"`
			Subject    string   `json:"subject"`
		} `json:"commonHeaders"`
	} `json:"mail"`
	Receipt struct {
		Timestamp            time.Time `json:"timestamp"`
		ProcessingTimeMillis int       `json:"processingTimeMillis"`
		Recipients           []string  `json:"recipients"`
		SpamVerdict          struct {
			Status string `json:"status"`
		} `json:"spamVerdict"`
		VirusVerdict struct {
			Status string `json:"status"`
		} `json:"virusVerdict"`
		SpfVerdict struct {
			Status string `json:"status"`
		} `json:"spfVerdict"`
		DkimVerdict struct {
			Status string `json:"status"`
		} `json:"dkimVerdict"`
		DmarcVerdict struct {
			Status string `json:"status"`
		} `json:"dmarcVerdict"`
		Action struct {
			Type     string `json:"type"`
			TopicArn string `json:"topicArn"`
			Encoding string `json:"encoding"`
		} `json:"action"`
	} `json:"receipt"`
	Content string `json:"content"`
}

func (m FallbackMail) SendMail(fromName, fromEmail string, toName string, toEmail []string, subject string, body string) (*string, error) {
	log.Printf("internal.platform.integration.email.fallback request - domain:%s  from: %s\n", m.Domain, fromEmail)
	resMsg, err := send()
	log.Printf("internal.platform.integration.email.fallback response - resMsg:%s  id: %s err:%v\n", resMsg, *resMsg.MessageId, err)
	return resMsg.MessageId, err
}

func (m FallbackMail) Watch(topicName string) (string, error) {
	return "", nil
}

func (m FallbackMail) Stop(emailAddress string) error {
	return nil
}

func send() (*ses.SendEmailOutput, error) {
	// Create a new session in the us-west-2 region.
	// Replace us-west-2 with the AWS Region you're using for Amazon SES.
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(AccessKey, SecretKey, "")},
	)

	// Create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(HtmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(CharSet),
					Data:    aws.String(TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(CharSet),
				Data:    aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	result, err := svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

	}

	fmt.Println("Email Sent to address: " + Recipient)
	fmt.Println(result)
	return result, err
}
