# Warnings and tips for developing Plugins
- if your plugin node is a pure Function,there is no need for using running mutex of the runner
- if your plugin needs to start a goroutine,then the mutex is a must