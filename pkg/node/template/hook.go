package template

type Hook interface {
	OnRegisterConflictRule(mgr ConflictManager)
}
