package sfnactivityworker

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/panjf2000/ants/v2"
)

type activityWorker struct {
	pool *ants.PoolWithFunc
}

type activityWorkerInput struct {
	Id                 string
	activityTaskOutput *sfn.GetActivityTaskOutput
}

func newActivityWorker[I, O any](act *activity[I, O]) (*activityWorker, error) {

	activityWorkerHandler := func(activityWorkerInput activityWorkerInput) bool {
		sfnActivityOut := activityWorkerInput.activityTaskOutput
		var activityInput I
		if err := json.Unmarshal([]byte(*sfnActivityOut.Input), &activityInput); err != nil {
			act.logger.WithValues("input", activityInput, "sfninput", sfnActivityOut.Input).Error(err, "An error occurred while marshalling sfn activity input to struct")
			panic(fmt.Errorf("Unable to marshall activity input to struct; %w", err))
		}
		activityOutput, activityErr := act.handler(act.ctx, activityInput)
		if activityErr != nil {
			if _, sendErr := act.sfnApi.SendTaskFailure(&sfn.SendTaskFailureInput{
				Cause:     aws.String(activityErr.Error()),
				Error:     aws.String(activityErr.Error()),
				TaskToken: sfnActivityOut.TaskToken,
			}); sendErr != nil {
				act.logger.WithValues("input", activityInput, "output", activityErr).Error(sendErr, "An error occurred while reporting failure to SFN!")
				panic(sendErr)
			}
		} else if outputJson, marshallErr := json.Marshal(activityOutput); marshallErr == nil {
			outputJsonString := string(outputJson)
			if _, sendErr := act.sfnApi.SendTaskSuccess(&sfn.SendTaskSuccessInput{
				Output:    &outputJsonString,
				TaskToken: sfnActivityOut.TaskToken,
			}); sendErr != nil {
				act.logger.WithValues("input", activityInput, "output", activityErr).Error(sendErr, "An error occurred while reporting success to SFN!")
				panic(sendErr)
			}
		} else {
			act.logger.WithValues("input", activityInput, "output", activityOutput).Error(marshallErr, "Error while sending task success request to SFN!")
		}
		return true
	}

	activityWorkerFunc := func(activityWorkerInputInterface interface{}) {
		if activityWorkerInput, ok := activityWorkerInputInterface.(activityWorkerInput); ok {
			resultSource := make(chan bool, 1)
			go func() {
				resultSource <- activityWorkerHandler(activityWorkerInput)
				close(resultSource)
			}()
			for {
				select {
				case <-resultSource:
					return
				case <-time.After(act.heartBeatDelay):
					if submitErr := act.heartbeatWorker.Submit(activityWorkerInput); submitErr != nil {
						act.logger.Error(submitErr, "Error while sending Heartbeat")
					}
				}
			}
		}
		panic(errors.New("Invalid Input From Poller Worker"))
	}

	opts := []ants.Option{
		ants.WithLogger(act.logger),
		ants.WithPanicHandler(act.panicHandler),
		ants.WithNonblocking(false),
		ants.WithMaxBlockingTasks(act.workers),
	}
	pool, err := ants.NewPoolWithFunc(act.workers, activityWorkerFunc, opts...)
	return &activityWorker{pool: pool}, err
}

func (actworker *activityWorker) Tune(size int) {
	actworker.pool.Tune(size)
}

func (actworker *activityWorker) Start() {
	actworker.pool.Reboot()
}

func (actworker *activityWorker) Stop() {
	actworker.pool.Release()
}

func (actworker *activityWorker) Submit(in interface{}) error {
	return actworker.pool.Invoke(in)
}
