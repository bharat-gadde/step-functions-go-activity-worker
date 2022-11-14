package sfnactivityworker

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/panjf2000/ants/v2"
)

type pollerWorker struct {
	pool *ants.PoolWithFunc
}

type pollerWorkerInput struct {
	Id string
}

func newPollerWorker[I, O any](act *activity[I, O]) (*pollerWorker, error) {
	pollerWorkerFunc := func(pollerWorkerInputInterface interface{}) {
		if pollerWorkerInput, ok := pollerWorkerInputInterface.(pollerWorkerInput); ok {
			if output, fetchErr := act.sfnApi.GetActivityTask(&sfn.GetActivityTaskInput{
				ActivityArn: aws.String(act.arn),
				WorkerName:  aws.String(pollerWorkerInput.Id),
			}); fetchErr != nil {
				act.logger.WithValues("WorkerId", pollerWorkerInput.Id).Error(fetchErr, "Error getting activity task")
				return
			} else if output.TaskToken == nil {
				return
			} else {
				if submitErr := act.activityWorker.Submit(activityWorkerInput{Id: pollerWorkerInput.Id, activityTaskOutput: output}); submitErr != nil {
					act.logger.WithValues("WorkerId", pollerWorkerInput.Id).Error(submitErr, "Error while invoking Worker")
				}
			}
		} else {
			panic(errors.New("Invalid Worker Id"))
		}
	}

	opts := []ants.Option{
		ants.WithLogger(act.logger),
		ants.WithPanicHandler(act.panicHandler),
		ants.WithNonblocking(false),
		ants.WithMaxBlockingTasks(act.pollers),
	}
	pool, err := ants.NewPoolWithFunc(act.pollers, pollerWorkerFunc, opts...)
	return &pollerWorker{pool: pool}, err
}

func (pworker *pollerWorker) Tune(size int) {
	pworker.pool.Tune(size)
}

func (pworker *pollerWorker) Start() {
	pworker.pool.Reboot()
}

func (pworker *pollerWorker) Stop() {
	pworker.pool.Release()
}

func (pworker *pollerWorker) Submit(pollerWorkerInput interface{}) error {
	return pworker.pool.Invoke(pollerWorkerInput)
}
