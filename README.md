[![Go Report Card](https://goreportcard.com/badge/github.com/bharat-gadde/step-functions-go-activity-worker)](https://goreportcard.com/report/github.com/bharat-gadde/step-functions-go-activity-worker)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/bharat-gadde/step-functions-go-activity-worker)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/bharat-gadde/step-functions-go-activity-worker?include_prereleases)

# step-functions-go-activity-worker

## Step Functions Activity Worker for Go

*Inspired from [step-functions-ruby-activity-worker](https://github.com/aws-samples/step-functions-ruby-activity-worker)*

Activity worker for polling, perform task and respond to step functions.

Abstracts the step function specifc logic from your code.

### Setting up the code

```go

// Custom logic
var activityHandler = func(ctx context.Context, act ActivityInput) (ActivityOutput, error) {
    fmt.Println(act.TransactionId)
    return ActivityOutput{TransactionID: act.TransactionId}, nil
}

// Create Activity
activity := sfnactivityworker.Activity(
    "arn:aws:states:ap-south-1:XXXXX:activity:XXXXX",
    activityHandler,
)
// Start the Activity
activity.Start()
// activityWorker.Stop() // to stop
```

### Tune during runtime

```go
activity.TunePollers(100)
activity.TuneWorkers(100000)
```

### Options available

```go
// Overrides the default SFN (Created with default session)
func WithSFN(sfn sfniface.SFNAPI) Options {
    return optionFunc(func(cfg config) config {
        cfg.sfnApi = sfn
        return cfg
    })
}

// Overrides the default HeartBeatDelay (60 Seconds)
func WithHeartBeatDelay(delay time.Duration) Options {
    return optionFunc(func(cfg config) config {
        cfg.heartBeatDelay = delay
        return cfg
    })
}

// Overrides the default WorkersCount (3)
func WithWorkersCount(workers int) Options {
    return optionFunc(func(cfg config) config {
        cfg.workers = workers
        return cfg
    })
}

// Overrides the default PollersCount (3)
func WithPollersCount(pollers int) Options {
    return optionFunc(func(cfg config) config {
        cfg.pollers = pollers
        return cfg
    })
}

// Overrides the default Logger (std)
func WithLogger(logger Logger) Options {
    return optionFunc(func(cfg config) config {
        cfg.logger = logger
        return cfg
    })
}

// Polling and Activity Handler can be bound to context
func WithContext(ctx context.Context) Options {
    return optionFunc(func(cfg config) config {
        cfg.ctx = ctx
        return cfg
    })
}

// Overrides the default Microsecond interval between consecutive Polls
func WithPollInterval(pollInterval time.Duration) Options {
    return optionFunc(func(cfg config) config {
        cfg.pollInterval = pollInterval
        return cfg
    })
}

```
