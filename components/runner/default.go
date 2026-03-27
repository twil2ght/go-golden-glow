package runner

import (
	"goldenglow/container"
	"goldenglow/node/template"
)

var (
	runnerInstance = New(container.DefaultFactory(), template.DefaultCore())
)

func DefaultRunner() Instance {
	return runnerInstance
}
