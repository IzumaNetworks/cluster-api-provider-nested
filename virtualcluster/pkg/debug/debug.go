//go:build vcdebugout
// +build vcdebugout

package debug

import (
	logr "github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DebugLogger struct {
	self logr.Logger
}

func DebugLoggerWithName(name string) *DebugLogger {
	return &DebugLogger{self: logf.Log.WithName("DEBUG-VC<" + name + ">")}
}

func (dl *DebugLogger) Info(msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = logf.Log.WithName("DEBUG-VC<unknown>")
	}
	dl.self.Info(msg, keysAndValues...)
}

func (dl *DebugLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = logf.Log.WithName("DEBUG-VC<unknown>")
	}
	dl.self.Error(err, msg, keysAndValues...)
}

func IfErr(err error) string {
	if err != nil {
		return err.Error()
	}
	return "[no err]"
}

func VcDoDebug(f func()) {
	f()
}
