package runner

import (
	"goldenglow/container"
	"goldenglow/node/template"
	"goldenglow/pkg/log"
)

var (
	runnerInstance = New(&log.Base{}, container.DefaultFactory(), template.DefaultCore())
)

func DefaultRunner() Instance {
	return runnerInstance
}
