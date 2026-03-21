package log

import "fmt"

type myErr struct {
	msg string
}

func NewErr(msg string) error {
	return &myErr{
		msg: msg,
	}
}
func (es *myErr) Error() string {
	return es.msg
}
func EmptyStrErr() error {
	return NewErr("empty str")
}
func NilErr() error {
	return NewErr("nil")
}
func LengthErr() error {
	return NewErr("len==0")
}
func NotFound(s string) error {
	return NewErr(fmt.Sprintf("%s Not Found", s))
}
func NotExist(field, got string) error {
	return fmt.Errorf("%s: %s doesn't exist", field, got)
}
