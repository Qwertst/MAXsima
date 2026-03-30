package ui

import (
	"bufio"
	"fmt"
	"io"

	"github.com/aydreq/maxsima/internal/model"
)

type UIRenderer interface {
	DisplayMessage(msg model.Message)
	ReadInput() (string, error)
}

type ConsoleUI struct {
	writer  io.Writer
	scanner *bufio.Scanner
}

func NewConsoleUI(w io.Writer, r io.Reader) *ConsoleUI {
	return &ConsoleUI{
		writer:  w,
		scanner: bufio.NewScanner(r),
	}
}

func (c *ConsoleUI) DisplayMessage(msg model.Message) {
	fmt.Fprintln(c.writer, msg.Format())
}

func (c *ConsoleUI) ReadInput() (string, error) {
	if c.scanner.Scan() {
		return c.scanner.Text(), nil
	}
	if err := c.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}
