package handler

type ExecuteHook interface {
	OnRegisterExecutor(reg Executor[ExecuteHandler])
}
type CheckHook interface {
	OnRegisterChecker(reg Executor[CheckHandler])
}
type ExtractHook interface {
	OnRegisterExtractor(reg Executor[ExtractorHandler])
}
