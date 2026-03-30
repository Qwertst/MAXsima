package chat

import (
	"fmt"
	"io"
	"time"

	"github.com/aydreq/maxsima/internal/model"
	"github.com/aydreq/maxsima/internal/ui"
)

// Message is a re-export of model.Message for use within the chat package tests.
type Message = model.Message

type Manager struct {
	username string
	ui       ui.UIRenderer
	session  *Session
	done     chan struct{}
}

// NewManager creates a new Manager. NewChatManager is an alias for compatibility.
func NewManager(username string, renderer ui.UIRenderer) *Manager {
	return &Manager{
		username: username,
		ui:       renderer,
		done:     make(chan struct{}),
	}
}

// NewChatManager is an alias for NewManager, used by tests.
func NewChatManager(renderer ui.UIRenderer, username string) *Manager {
	return NewManager(username, renderer)
}

// GetUsername returns the username of this manager.
func (m *Manager) GetUsername() string {
	return m.username
}

func (m *Manager) StartSession(sender MessageSender, receiver MessageReceiver) error {
	m.session = NewSession(sender, receiver, m.username)

	go m.handleIncoming()
	go m.handleOutgoing()

	return nil
}

func (m *Manager) Wait() {
	<-m.done
}

// StopSession stops the active session and returns nil.
func (m *Manager) StopSession() error {
	if m.session != nil {
		m.session.Close()
	}
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	return nil
}

func (m *Manager) handleIncoming() {
	for m.session.IsActive() {
		msg, err := m.session.receiver.Receive()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("receive error: %v\n", err)
			}
			m.StopSession()
			return
		}
		m.ui.DisplayMessage(msg)
	}
}

func (m *Manager) handleOutgoing() {
	for m.session.IsActive() {
		text, err := m.ui.ReadInput()
		if err != nil {
			if err == io.EOF {
				m.StopSession()
			}
			return
		}
		if text == "" {
			continue
		}
		msg := model.Message{
			SenderName: m.username,
			Timestamp:  time.Now(),
			Text:       text,
		}
		if err := m.session.sender.Send(msg); err != nil {
			fmt.Printf("send error: %v\n", err)
			m.StopSession()
			return
		}
	}
}
