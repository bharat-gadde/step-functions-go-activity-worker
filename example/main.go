package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	sfnactivityworker "github.com/bharat-gadde/step-functions-go-activity-worker"
)

type ActivityInput struct {
	TransactionId string `json:"transactionId"`
}
type ActivityOutput struct {
	TransactionID string `json:"transactionId"`
}

var activityHandler = func(ctx context.Context, act ActivityInput) (ActivityOutput, error) {
	return ActivityOutput{TransactionID: act.TransactionId}, nil
}

func main() {

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	sfnAPI := sfn.New(awsSession)

	activity := sfnactivityworker.Activity(
		"arn:aws:states:ap-south-1:732165046977:activity:PG2DevUpdateTransactionStatusActivity",
		activityHandler,
		sfnactivityworker.WithSFN(sfnAPI),
		sfnactivityworker.WithHeartBeatDelay(3*time.Second),
	)
	activity.Start()
	// activityWorker.Stop() // to stop
	kill := make(chan struct{})
	// activity.TunePollers(100)
	// activity.TuneWorkers(100000)
	<-kill
}
