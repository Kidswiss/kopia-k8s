package logger

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync/atomic"

	"github.com/go-logr/logr"
)

// ContextKey is used to reference the logger in a given context
type ContextKey struct{}

// AppLogger retrieves the application-wide logger instance from the context.Context.
func AppLogger(c context.Context) logr.Logger {
	return c.Value(ContextKey{}).(*atomic.Value).Load().(logr.Logger)
}

type outFunc func(string)

// New creates a writer which directly writes to the given logger function.
func New(out outFunc) io.Writer {
	return &writer{out}
}

// NewInfoWriter creates a writer with the name "stdout" which directly writes to the given logger using info level.
// It ensures that each line is handled separately. This avoids mangled lines when parsing
// JSON outputs.
func NewInfoWriter(l logr.Logger) io.Writer {
	return New((&LogInfoPrinter{l}).out)
}

// NewErrorWriter creates a writer with the name "stderr" which directly writes to the given logger using info level.
// It ensures that each line is handled seperately. This avoids mangled lines when parsing
// JSON outputs.
func NewErrorWriter(l logr.Logger) io.Writer {
	return New((&LogErrPrinter{l}).out)
}

type writer struct {
	out outFunc
}

func (w writer) Write(p []byte) (int, error) {

	scanner := bufio.NewScanner(bytes.NewReader(p))

	for scanner.Scan() {
		w.out(scanner.Text())
	}

	return len(p), nil
}

type LogInfoPrinter struct {
	log logr.Logger
}

func (l *LogInfoPrinter) out(s string) {
	l.log.WithName("stdout").Info(s)
}

type LogErrPrinter struct {
	Log logr.Logger
}

func (l *LogErrPrinter) out(s string) {
	l.Log.WithName("stderr").Info(s)
}
