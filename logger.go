package sfnactivityworker

import (
	"fmt"

	"github.com/go-logr/logr"
)

type Logger struct {
	logr.Logger
}

func (loggr Logger) Printf(format string, args ...interface{}) {
	loggr.WithValues(args).Info(fmt.Sprintf(format, args...))
}
