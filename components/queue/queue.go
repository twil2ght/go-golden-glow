package queue

type Queue interface {
	In(string) error
	Out() <-chan string
}
