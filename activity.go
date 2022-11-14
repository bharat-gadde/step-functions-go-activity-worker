package sfnactivityworker

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/go-logr/stdr"
	"github.com/google/uuid"
)

const (
	stopped int64 = 0
	running int64 = 1
)

type config struct {
	sfnApi         sfniface.SFNAPI
	heartBeatDelay time.Duration
	workers        int
	pollers        int
	maxRetries     int
	pollInterval   time.Duration
	logger         Logger
	panicHandler   func(interface{})
	ctx            context.Context
}
type Handler[I, O any] func(context.Context, I) (O, error)

type activity[I, O any] struct {
	config
	arn             string
	handler         Handler[I, O]
	activityWorker  worker
	pollerWorker    worker
	heartbeatWorker worker
	state           int64
}

var defaultConfig = func() config {
	c := config{
		heartBeatDelay: 60 * time.Second,
		workers:        3,
		pollers:        3,
		maxRetries:     3,
		pollInterval:   time.Microsecond,
		logger:         Logger{stdr.NewWithOptions(stdlog.New(os.Stdout, "", stdlog.LstdFlags), stdr.Options{LogCaller: stdr.All})},
		ctx:            context.Background(),
	}
	if sess, err := session.NewSession(); err == nil {
		c.sfnApi = sfn.New(sess, aws.NewConfig().WithMaxRetries(c.maxRetries))
	}
	return c
}

func Activity[I, O any](arn string, handler Handler[I, O], options ...Options) *activity[I, O] {
	config := defaultConfig()
	for _, o := range options {
		config = o.apply(config)
	}
	return &activity[I, O]{
		arn:     arn,
		config:  config,
		handler: handler,
		state:   stopped,
	}
}

// Start starts the activity
func (a *activity[I, O]) Start() error {
	if atomic.CompareAndSwapInt64(&a.state, stopped, running) {
		var errors []error
		if heartbeatWorker, heartbeatWorkerErr := newHeartbeatWorker(a); heartbeatWorkerErr == nil {
			a.heartbeatWorker = heartbeatWorker
		} else {
			errors = append(errors, heartbeatWorkerErr)
			a.logger.Error(heartbeatWorkerErr, "Error while starting heartbeat worker")
		}

		if activityWorker, activityWorkerErr := newActivityWorker(a); activityWorkerErr == nil {
			a.activityWorker = activityWorker
		} else {
			errors = append(errors, activityWorkerErr)
			a.logger.Error(activityWorkerErr, "Error while starting heartbeat worker")
		}

		if pollerWorker, pollerWorkerErr := newPollerWorker(a); pollerWorkerErr == nil {
			a.pollerWorker = pollerWorker
		} else {
			errors = append(errors, pollerWorkerErr)
			a.logger.Error(pollerWorkerErr, "Error while starting heartbeat worker")
		}

		if len(errors) > 0 {
			atomic.StoreInt64(&a.state, stopped)
			return fmt.Errorf("Error while starting activity")
		}
		a.pollerWorker.Start()
		a.activityWorker.Start()
		a.heartbeatWorker.Start()

		go func() {
			pollIntervalTicker := time.NewTicker(a.pollInterval)
			for a.state == running {
				select {
				case <-a.ctx.Done():
					break
				case <-pollIntervalTicker.C:
					if submitErr := a.pollerWorker.Submit(pollerWorkerInput{Id: uuid.NewString()}); submitErr != nil {
						a.logger.Error(submitErr, "Error while invoking a poller")
					}
				default:
					continue
				}
			}
		}()

	}
	return nil
}

// Stop the activity
func (a *activity[I, O]) Stop() error {
	if atomic.CompareAndSwapInt64(&a.state, running, stopped) {
		a.pollerWorker.Stop()
		a.activityWorker.Stop()
		a.heartbeatWorker.Stop()
	}
	return nil
}

// Tune Number of pollers
func (a *activity[I, O]) TunePollers(size int) {
	a.pollerWorker.Tune(size)
}

// Tune's number of workers
func (a *activity[I, O]) TuneWorkers(size int) {
	a.activityWorker.Tune(size)
}
