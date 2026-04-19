package messageQueue

type MsgQueueHook interface {
	OnRegisterMsgProvider(reg Manager)
}
