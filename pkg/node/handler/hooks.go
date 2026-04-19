package handler

type ExecuteHook interface {
	OnRegisterExecutor(reg Executor[ExecuteHandler])
}
type CheckHook interface {
	OnCheckExecutor(reg Executor[CheckHandler])
}
type ExtractHook interface {
	OnExtractExecutor(reg Executor[ExtractorHandler])
}
