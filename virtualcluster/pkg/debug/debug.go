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

var debugLogger = &DebugLogger{
	self: logf.Log.WithName("DEBUG-VC"),
}

func DebugLoggerWithName(name string) DebugLogger {
	return DebugLogger{self: debugLogger.self.WithName(name)}
}

func Info(msg string, keysAndValues ...interface{}) {
	debugLogger.self.Info(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	debugLogger.self.Error(err, msg, keysAndValues...)
}

func (dl DebugLogger) Info(msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = debugLogger.self.WithName("unknown")
	}
	dl.self.Info(msg, keysAndValues...)
}

func (dl DebugLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = debugLogger.self.WithName("unknown")
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
