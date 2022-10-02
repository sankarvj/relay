package job

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

func (j *Job) Stream(message *stream.Message) error {
	build := util.ExpvarGet("build")
	var err error
	if build != "prod" && build != "stage" {
		err = j.Post(message)
	} else {
		err = queueSQS(message)
	}
	if err != nil {
		log.Println("***> unexpected error occurred when queing message to SQS", err)
		return err
	}

	return nil
}

func queueSQS(message *stream.Message) error {
	region := util.ExpvarGet("aws_region")
	queueURL := util.ExpvarGet("aws_worker_sqs_url")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(region), //TODO: don't hardcode take this param from the ENV
		Endpoint: aws.String(fmt.Sprintf("https://sqs.%s.amazonaws.com", region)),
	})
	if err != nil {
		return err
	}
	svc := sqs.New(sess)

	// Make message JSON
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.ID),
			},
			"Action": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.TypeStr()),
			},
		},
		MessageBody: aws.String(string(msg)),
		QueueUrl:    aws.String(queueURL),
	})
	return err
}
