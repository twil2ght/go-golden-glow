package datagen

type Hook interface {
	OnRegisterDataGen(gen Generator)
}
