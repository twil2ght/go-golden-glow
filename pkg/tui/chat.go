package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Message struct {
	Sender string // "You" or speaker name (e.g. "Susie")
	Text   string
	Time   time.Time
}

type Chat struct {
	out     io.Writer
	in      *bufio.Scanner
	inputCh chan string
	msgCh   chan Message
}

func NewChat() *Chat {
	return &Chat{
		out:     os.Stdout,
		in:      bufio.NewScanner(os.Stdin),
		inputCh: make(chan string, 1),
		msgCh:   make(chan Message, 64),
	}
}

// Display queues a message to be shown in the chat.
func (c *Chat) Display(msg Message) {
	c.msgCh <- msg
}

// InputCh returns the channel that receives user input strings.
func (c *Chat) InputCh() <-chan string {
	return c.inputCh
}

// Start runs the chat UI in the foreground. It blocks until ctx is done.
// It draws the header, then loops: displays pending messages and prompts for input.
func (c *Chat) Start(ctx context.Context) {
	c.renderHeader()

	// Print initial prompt
	c.printPrompt()

	// Channel to signal that a line of input is available
	type scanResult struct {
		text string
		ok   bool
	}
	scanCh := make(chan scanResult, 1)
	go func() {
		for {
			if c.in.Scan() {
				scanCh <- scanResult{text: c.in.Text(), ok: true}
			} else {
				scanCh <- scanResult{ok: false}
				return
			}
		}
	}()

	for {
		// Wait for either a display message or user input
		select {
		case <-ctx.Done():
			return
		case msg := <-c.msgCh:
			c.printMessage(msg)
			c.printPrompt()
		case res := <-scanCh:
			if !res.ok {
				return
			}
			input := strings.TrimSpace(res.text)
			if input == "" {
				c.printPrompt()
				continue
			}
			// Show user's message
			c.printMessage(Message{Sender: "You", Text: input, Time: time.Now()})
			// Send to pipeline
			c.inputCh <- input
			c.printPrompt()
		}
	}
}

var (
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorDim    = "\033[2m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
)

func (c *Chat) renderHeader() {
	line := "────────────────────────────────────────────────"
	fmt.Fprintf(c.out, "%s┌%s┐%s\n", colorCyan, line, colorReset)
	fmt.Fprintf(c.out, "%s│%s  %s◆  Golden Glow  ◆%s%s                   %s│%s\n",
		colorCyan, colorReset, colorBold+colorCyan, colorReset, colorDim, colorCyan, colorReset)
	fmt.Fprintf(c.out, "%s└%s┘%s\n", colorCyan, line, colorReset)
	fmt.Fprintln(c.out)
}

func (c *Chat) printMessage(msg Message) {
	ts := msg.Time.Format("15:04:05")
	var senderColor string
	switch msg.Sender {
	case "You":
		senderColor = colorYellow
	default:
		senderColor = colorCyan
	}
	fmt.Fprintf(c.out, "  %s%s%s %s(%s)%s\n", senderColor, msg.Sender, colorReset, colorDim, ts, colorReset)
	for _, line := range strings.Split(msg.Text, "\n") {
		fmt.Fprintf(c.out, "  %s%s\n", colorGreen, line)
	}
	fmt.Fprintln(c.out)
}

func (c *Chat) printPrompt() {
	fmt.Fprintf(c.out, "%s  > %s", colorYellow, colorReset)
}
