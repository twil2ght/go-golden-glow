package queue

type Queue interface {
	In(string) error
	Out() <-chan string
}
type queue struct {
	ch chan string
}

func (q *queue) In(s string) error {
	if s == "" {
		return nil
	}
	q.ch <- s
	return nil
}

func (q *queue) Out() <-chan string {
	return q.ch
}
func NewQueue() Queue {
	return &queue{
		ch: make(chan string),
	}
}
