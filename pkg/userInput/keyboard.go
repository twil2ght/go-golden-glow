package userInput

import (
	"bufio"
	"context"
	"fmt"
	"goldenglow/pkg/messageQueue"
	"os"
)

type Keyboard struct {
	queue messageQueue.Interface
}

func (k *Keyboard) Start(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	inputCh := make(chan string)
	errChan := make(chan error)
	go func() {
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				errChan <- scanner.Err()
				break
			}
			inputCh <- scanner.Text()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-errChan:
			return
		case input, ok := <-inputCh:
			if !ok {
				return
			}
			k.queue.Add(input)
		}
	}
}
func NewKeyboard(queue messageQueue.Interface) *Keyboard {
	return &Keyboard{
		queue: queue,
	}
}
