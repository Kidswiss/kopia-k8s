package k8s

import (
	"github.com/go-logr/logr"
)

type execLogger struct {
	log       logr.Logger
	namespace string
	podname   string
}

func (e *execLogger) execStdout(line string) {
	e.log.Info(line, "podname", e.podname, "namespace", e.namespace)
}

func (e *execLogger) execStderr(line string) {
	e.log.Error(nil, line, "podname", e.podname, "namespace", e.namespace)
}
