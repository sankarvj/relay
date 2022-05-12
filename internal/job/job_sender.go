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
	log.Println("build -- ", build)
	var err error
	if build == nil || build.String() == `"develop"` {
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
		Region: aws.String("us-east-1"),
	},
	)

	if err != nil {
		return err
	}

	// Create an SES session.
	svc := sqs.New(sess)

	// Make message JSON
	msg, err := json.Marshal(message)
	if err != nil {
		log.Println("***>***> EventItemCreated: unexpected/unhandled error occurred when sending the job to SQS. error:", err)
		return err
	}

	_, err = svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Title": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Job"),
			},
			"Action": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Item Create"),
			},
		},
		MessageBody: aws.String(string(msg)),
		QueueUrl:    aws.String(queueURL),
	})
	return err
}
