package model

import (
	"fmt"
	"time"
)

type Message struct {
	SenderName string
	Timestamp  time.Time
	Text       string
}

func (m Message) Format() string {
	return fmt.Sprintf("[%s] %s: %s", m.Timestamp.Format("2006-01-02 15:04:05"), m.SenderName, m.Text)
}
