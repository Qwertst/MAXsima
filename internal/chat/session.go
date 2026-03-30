package chat

import (
	"sync"

	"github.com/aydreq/maxsima/internal/model"
)

type MessageSender interface {
	Send(msg model.Message) error
}

type MessageReceiver interface {
	Receive() (model.Message, error)
}

type Session struct {
	sender   MessageSender
	receiver MessageReceiver
	username string
	active   bool
	mu       sync.Mutex
}

func NewSession(sender MessageSender, receiver MessageReceiver, username string) *Session {
	return &Session{
		sender:   sender,
		receiver: receiver,
		username: username,
		active:   true,
	}
}

func (s *Session) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = false
}
