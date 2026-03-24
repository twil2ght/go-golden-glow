package components

type Message interface {
	Value() string
	Tag() string
}
