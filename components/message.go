package components

type msg struct {
	value, tag string
}

func (m *msg) Value() string {
	return m.value
}

func (m *msg) Tag() string {
	return m.tag
}
func NewMsg(value, tag string) Message {
	return &msg{value, tag}
}
