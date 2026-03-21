package goldenglow

type Source interface {
	C() <-chan Message
}
type Queue interface {
	In(string) error
	Out() <-chan string
}
type ProcessorHandler func(msg string) string
type Processor interface {
	Do(message Message) string
	Register(tag string, handler ProcessorHandler) error
}
type Message interface {
	Value() string
	Tag() string
}
