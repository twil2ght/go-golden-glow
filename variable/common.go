package variable

func Is(str string) bool {
	return VarReg.MatchString(str)
}
