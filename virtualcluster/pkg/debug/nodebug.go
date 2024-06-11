//go:build !vcdebugout
// +build !vcdebugout

package debug

type DebugLogger struct {
	// self logr.Logger
}

func DebugLoggerWithName(name string) *DebugLogger {
	return &DebugLogger{}
}

func (dl *DebugLogger) Info(msg string, keysAndValues ...interface{}) {
	// dl.self.Info(msg, keysAndValues...)
}

func (dl *DebugLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	// dl.self.Error(err, msg, keysAndValues...)
}

func IfErr(err error) string {
	return ""
}

func VcDoDebug(f func()) {
	// f()
}
