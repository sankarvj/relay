package job

import (
	"encoding/json"
	"expvar"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
)

const (
	queueURL = "https://sqs.us-east-1.amazonaws.com/191933142379/awseb-e-yhaas2daqe-stack-AWSEBWorkerQueue-io2v9Ty1Rlxh"
)

func (j *Job) Stream(message *stream.Message) error {
	build := expvar.Get("build")
	var err error
	if build == nil || build.String() == `"dev"` {
		err = j.Post(message)
	} else {
		err = queueSQS(message)
	}
	if err != nil {
		return err
	}

	return nil
}

func queueSQS(message *stream.Message) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), //TODO: don't hardcode take this param from the ENV
	},
	)

	if err != nil {
		return err
	}

	// Create an SQS session.
	svc := sqs.New(sess)

	// Make message JSON
	msg, err := json.Marshal(message)
	if err != nil {
		log.Println("***>***> queueSQS: unexpected/unhandled error occurred when sending the job to SQS. error:", err)
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
