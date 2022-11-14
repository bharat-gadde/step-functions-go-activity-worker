package sfnactivityworker

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/panjf2000/ants/v2"
)

type heartbeatWorker struct {
	pool *ants.PoolWithFunc
}

type heartbeatWorkerInput activityWorkerInput

func newHeartbeatWorker[I, O any](act *activity[I, O]) (*heartbeatWorker, error) {
	heartbeatWorkerFunc := func(heartbeatWorkerInputInterface interface{}) {
		if heartbeatWorkerInput, ok := heartbeatWorkerInputInterface.(heartbeatWorkerInput); ok {
			act.logger.Info("Sending Heartbeat")
			sfnActivityOut := heartbeatWorkerInput.activityTaskOutput
			if _, sendErr := act.sfnApi.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
				TaskToken: sfnActivityOut.TaskToken,
			}); sendErr != nil {
				act.logger.WithValues("input", sfnActivityOut.Input).Error(sendErr, "Error while sending HeartBeat to SFN!")
			}
			act.logger.Info("Sent Heartbeat")
		} else {
			panic(errors.New("Invalid Input From Poller Worker"))
		}
	}

	opts := []ants.Option{
		ants.WithLogger(act.logger),
		ants.WithPanicHandler(act.panicHandler),
	}
	pool, err := ants.NewPoolWithFunc(-1, heartbeatWorkerFunc, opts...)
	return &heartbeatWorker{pool: pool}, err
}

func (htworker *heartbeatWorker) Tune(size int) {
	htworker.pool.Tune(size)
}

func (htworker *heartbeatWorker) Start() {
	htworker.pool.Reboot()
}

func (htworker *heartbeatWorker) Stop() {
	htworker.pool.Release()
}

func (htworker *heartbeatWorker) Submit(in interface{}) error {
	return htworker.pool.Invoke(in)
}
