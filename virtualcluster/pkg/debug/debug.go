//go:build vcdebugout
// +build vcdebugout

package debug

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

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

// getShortFilePath returns the last 'depth' elements of the file path
func partOfPath(fullPath string, depth int) string {
	elements := strings.Split(fullPath, string(filepath.Separator))
	if len(elements) <= depth {
		return fullPath
	}
	return filepath.Join(elements[len(elements)-depth:]...)
}

func Info(msg string, keysAndValues ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fpath := partOfPath(file, 2)
	src := fmt.Sprintf("%s:%d", fpath, line)
	keysAndValues = append(keysAndValues, "SRC", src)
	debugLogger.self.Info(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fpath := partOfPath(file, 2)
	src := fmt.Sprintf("%s:%d", fpath, line)
	keysAndValues = append(keysAndValues, "SRC", src)

	debugLogger.self.Error(err, msg, keysAndValues...)
}

func (dl DebugLogger) Info(msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = debugLogger.self.WithName("unknown")
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fpath := partOfPath(file, 2)
	src := fmt.Sprintf("%s:%d", fpath, line)
	keysAndValues = append(keysAndValues, "SRC", src)

	dl.self.Info(msg, keysAndValues...)
}

func (dl DebugLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if dl.self == nil {
		dl.self = debugLogger.self.WithName("unknown")
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fpath := partOfPath(file, 2)
	src := fmt.Sprintf("%s:%d", fpath, line)
	keysAndValues = append(keysAndValues, "SRC", src)

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
