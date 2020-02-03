package aws

import "time"

//Subscription decribes the aws sns object
type Subscription struct {
	Type             string    `json:"Type"`
	MessageID        string    `json:"MessageId"`
	Token            string    `json:"Token"`
	TopicArn         string    `json:"TopicArn"`
	Subject          string    `json:"Subject"`
	Message          string    `json:"Message"`
	SubscribeURL     string    `json:"SubscribeURL"`
	Timestamp        time.Time `json:"Timestamp"`
	SignatureVersion string    `json:"SignatureVersion"`
	Signature        string    `json:"Signature"`
	SigningCertURL   string    `json:"SigningCertURL"`
	UnsubscribeURL   string    `json:"UnsubscribeURL"`
}

//Message descripbes sns message
type Message struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription"`
	AWSAccountID     string `json:"AWSAccountId"`
	NewStateValue    string `json:"NewStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	StateChangeTime  string `json:"StateChangeTime"`
	Region           string `json:"Region"`
	OldStateValue    string `json:"OldStateValue"`
	Trigger          struct {
		MetricName                       string      `json:"MetricName"`
		Namespace                        string      `json:"Namespace"`
		StatisticType                    string      `json:"StatisticType"`
		Statistic                        string      `json:"Statistic"`
		Unit                             interface{} `json:"Unit"`
		Dimensions                       []Dimension `json:"Dimensions"`
		Period                           int         `json:"Period"`
		EvaluationPeriods                int         `json:"EvaluationPeriods"`
		ComparisonOperator               string      `json:"ComparisonOperator"`
		Threshold                        float64     `json:"Threshold"`
		TreatMissingData                 string      `json:"TreatMissingData"`
		EvaluateLowSampleCountPercentile string      `json:"EvaluateLowSampleCountPercentile"`
	} `json:"Trigger"`
}

//Dimension describes sns key values
type Dimension struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}
