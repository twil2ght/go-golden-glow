package preprocessor

import "goldenglow/components"

type Handler func(msg string) string
type Instance interface {
	Preprocess(message components.Message) string
	Register(tag string, handler Handler) error
}
