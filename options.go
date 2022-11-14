package sfnactivityworker

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
)

type Options interface {
	apply(config) config
}

type optionFunc func(cfg config) config

func (fn optionFunc) apply(cfg config) config {
	return fn(cfg)
}

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
