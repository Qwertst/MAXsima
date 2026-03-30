package chat

import (
	"testing"
	"sync"
	"time"
)

func TestChatManagerStartSession(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	sender := &MockMessageSender{}
	receiver := &MockMessageReceiver{}

	err := mgr.StartSession(sender, receiver)

	if err != nil {
		t.Errorf("StartSession should not produce error, got: %v", err)
	}
}

func TestChatManagerReceivesIncomingMessages(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	incomingMessages := []Message{
		{
			SenderName: "Bob",
			Timestamp:  time.Now(),
			Text:       "Hello Alice!",
		},
		{
			SenderName: "Bob",
			Timestamp:  time.Now(),
			Text:       "How are you?",
		},
	}
	receiver := NewMockMessageReceiverWithMessages(incomingMessages)
	sender := &MockMessageSender{}

	mgr.StartSession(sender, receiver)

	time.Sleep(100 * time.Millisecond)

	msgs := ui.GetDisplayedMessages()
	if len(msgs) != 2 {
		t.Errorf("Expected 2 displayed messages, got %d", len(msgs))
	} else if msgs[0].SenderName != "Bob" || msgs[0].Text != "Hello Alice!" {
		t.Errorf("First message not displayed correctly: %v", msgs[0])
	}
}

func TestChatManagerSendsUserMessages(t *testing.T) {
	mUI := &MockUIRenderer{}
	mgr := NewChatManager(mUI, "Alice")

	sender := &MockMessageSender{}
	// Use a slow receiver so the session stays active long enough for
	// handleOutgoing to send both messages before the session ends.
	receiver := NewMockSlowMessageReceiver(500 * time.Millisecond)

	// Prime the UI with two messages before starting the session.
	mUI.SetInputs([]string{"Hello Bob!", "How are you?"})

	mgr.StartSession(sender, receiver)
	time.Sleep(200 * time.Millisecond)

	sent := sender.GetSentMessages()
	if len(sent) < 2 {
		t.Errorf("expected 2 sent messages, got %d", len(sent))
		return
	}
	if sent[0].Text != "Hello Bob!" || sent[0].SenderName != "Alice" {
		t.Errorf("unexpected first sent message: %+v", sent[0])
	}
	if sent[1].Text != "How are you?" {
		t.Errorf("unexpected second sent message: %+v", sent[1])
	}
}

func TestChatManagerSendErrorStopsSession(t *testing.T) {
	mUI := &MockUIRenderer{}
	mgr := NewChatManager(mUI, "Alice")

	failSender := &MockFailingSender{}
	receiver := NewMockMessageReceiverWithMessages([]Message{})
	mUI.SetInputs([]string{"trigger error"})

	mgr.StartSession(failSender, receiver)

	// Wait() should unblock after the send error stops the session.
	done := make(chan struct{})
	go func() { mgr.Wait(); close(done) }()

	select {
	case <-done:
		// session stopped due to send error — pass
	case <-time.After(2 * time.Second):
		t.Errorf("session did not stop after send error")
	}
}

func TestChatManagerWait(t *testing.T) {
	mUI := &MockUIRenderer{}
	mgr := NewChatManager(mUI, "Alice")

	// Receiver returns one message then EOF.
	receiver := NewMockMessageReceiverWithMessages([]Message{
		{SenderName: "Bob", Text: "hi", Timestamp: time.Now()},
	})
	sender := &MockMessageSender{}

	mgr.StartSession(sender, receiver)

	done := make(chan struct{})
	go func() { mgr.Wait(); close(done) }()

	select {
	case <-done:
		// Wait() returned after session ended — pass
	case <-time.After(2 * time.Second):
		t.Errorf("Wait() did not return after session ended")
	}
}

func TestChatManagerHandlesConnectionClosure(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	receiver := NewMockMessageReceiverWithError(ErrConnectionClosed)
	sender := &MockMessageSender{}

	err := mgr.StartSession(sender, receiver)

	if err != nil && err != ErrConnectionClosed {
		t.Errorf("StartSession should handle closed connection gracefully, got: %v", err)
	}
}

func TestChatManagerStopSession(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	sender := &MockMessageSender{}
	receiver := NewMockMessageReceiverWithMessages(make([]Message, 0))

	mgr.StartSession(sender, receiver)

	err := mgr.StopSession()

	if err != nil {
		t.Errorf("StopSession should not produce error, got: %v", err)
	}
}

func TestChatManagerBidirectionalMessaging(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	incomingMessages := []Message{
		{SenderName: "Bob", Text: "Hello!", Timestamp: time.Now()},
		{SenderName: "Bob", Text: "Goodbye!", Timestamp: time.Now()},
	}
	receiver := NewMockMessageReceiverWithMessages(incomingMessages)

	sender := &MockMessageSender{}

	userMessages := []string{"Hi Bob!", "See you!"}
	ui.SetInputs(userMessages)

	mgr.StartSession(sender, receiver)
	time.Sleep(100 * time.Millisecond)

	if len(ui.GetDisplayedMessages()) < len(incomingMessages) {
		t.Errorf("Not all incoming messages were displayed")
	}
}

func TestChatManagerMessageFormat(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Charlie")

	incomingMsg := Message{
		SenderName: "David",
		Timestamp:  time.Date(2026, 3, 30, 15, 45, 30, 0, time.UTC),
		Text:       "Test message",
	}
	receiver := NewMockMessageReceiverWithMessages([]Message{incomingMsg})
	sender := &MockMessageSender{}

	mgr.StartSession(sender, receiver)
	time.Sleep(50 * time.Millisecond)

	msgs := ui.GetDisplayedMessages()
	if len(msgs) > 0 {
		if msgs[0].SenderName != "David" || msgs[0].Text != "Test message" {
			t.Errorf("Message not formatted correctly")
		}
	}
}

func TestChatManagerPreservesUsername(t *testing.T) {
	ui := &MockUIRenderer{}
	username := "EdnaMode"
	mgr := NewChatManager(ui, username)

	retrievedUsername := mgr.GetUsername()

	if retrievedUsername != username {
		t.Errorf("ChatManager didn't preserve username: expected '%s', got '%s'", username, retrievedUsername)
	}
}

func TestChatManagerConcurrentReadWrite(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Frank")

	slowReceiver := NewMockSlowMessageReceiver(100 * time.Millisecond)
	sender := &MockMessageSender{}

	startTime := time.Now()
	mgr.StartSession(sender, slowReceiver)
	time.Sleep(50 * time.Millisecond)

	elapsed := time.Since(startTime)
	if elapsed > 200*time.Millisecond {
		t.Logf("Concurrent execution might have blocking issues: took %v", elapsed)
	}
}

type MockUIRenderer struct {
	displayedMessages []Message
	inputQueue        []string
	inputIndex        int
	mutex             sync.Mutex
}

func (m *MockUIRenderer) DisplayMessage(msg Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.displayedMessages = append(m.displayedMessages, msg)
}

// GetDisplayedMessages returns a safe copy of all displayed messages.
func (m *MockUIRenderer) GetDisplayedMessages() []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	out := make([]Message, len(m.displayedMessages))
	copy(out, m.displayedMessages)
	return out
}

func (m *MockUIRenderer) ReadInput() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.inputIndex < len(m.inputQueue) {
		result := m.inputQueue[m.inputIndex]
		m.inputIndex++
		return result, nil
	}
	return "", ErrNoInput
}

func (m *MockUIRenderer) SetInputs(inputs []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.inputQueue = inputs
	m.inputIndex = 0
}

type MockMessageSender struct {
	sentMessages []Message
	mutex        sync.Mutex
}

func (m *MockMessageSender) Send(msg Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

// GetSentMessages returns a safe copy of all sent messages.
func (m *MockMessageSender) GetSentMessages() []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	out := make([]Message, len(m.sentMessages))
	copy(out, m.sentMessages)
	return out
}

// MockFailingSender always returns an error from Send.
type MockFailingSender struct{}

func (m *MockFailingSender) Send(msg Message) error {
	return ErrConnectionClosed
}

type MockMessageReceiver struct {
	Messages []Message
	index    int
	mutex    sync.Mutex
}

func NewMockMessageReceiverWithMessages(msgs []Message) *MockMessageReceiver {
	return &MockMessageReceiver{
		Messages: msgs,
		index:    0,
	}
}

func (m *MockMessageReceiver) Receive() (Message, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.index < len(m.Messages) {
		result := m.Messages[m.index]
		m.index++
		return result, nil
	}
	return Message{}, ErrConnectionClosed
}

type MockMessageReceiverWithError struct {
	err error
}

func NewMockMessageReceiverWithError(err error) *MockMessageReceiverWithError {
	return &MockMessageReceiverWithError{err: err}
}

func (m *MockMessageReceiverWithError) Receive() (Message, error) {
	return Message{}, m.err
}

type MockSlowMessageReceiver struct {
	delay time.Duration
	count int
	mutex sync.Mutex
}

func NewMockSlowMessageReceiver(delay time.Duration) *MockSlowMessageReceiver {
	return &MockSlowMessageReceiver{delay: delay}
}

func (m *MockSlowMessageReceiver) Receive() (Message, error) {
	time.Sleep(m.delay)
	m.mutex.Lock()
	m.count++
	m.mutex.Unlock()
	if m.count > 3 {
		return Message{}, ErrConnectionClosed
	}
	return Message{SenderName: "SlowPeer", Text: "Slow message"}, nil
}

var (
	ErrConnectionClosed = &ChatError{Code: "CONNECTION_CLOSED", Message: "Connection closed"}
	ErrNoInput          = &ChatError{Code: "NO_INPUT", Message: "No input available"}
)

type ChatError struct {
	Code    string
	Message string
}

func (e *ChatError) Error() string {
	return e.Message
}
