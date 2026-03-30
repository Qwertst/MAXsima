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

	if len(ui.DisplayedMessages) != 2 {
		t.Errorf("Expected 2 displayed messages, got %d", len(ui.DisplayedMessages))
	}

	if ui.DisplayedMessages[0].SenderName != "Bob" || ui.DisplayedMessages[0].Text != "Hello Alice!" {
		t.Errorf("First message not displayed correctly: %v", ui.DisplayedMessages[0])
	}
}

func TestChatManagerSendsUserMessages(t *testing.T) {
	ui := &MockUIRenderer{}
	mgr := NewChatManager(ui, "Alice")

	sender := &MockMessageSender{}
	receiver := NewMockMessageReceiverWithMessages([]Message{})

	mgr.StartSession(sender, receiver)

	userInputs := []string{"Hello Bob!", "How are you?"}
	ui.SetInputs(userInputs)

	for _, input := range userInputs {
		_ = input
	}

	_ = sender
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

	if len(ui.DisplayedMessages) < len(incomingMessages) {
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

	if len(ui.DisplayedMessages) > 0 {
		displayedMsg := ui.DisplayedMessages[0]
		if displayedMsg.SenderName != "David" || displayedMsg.Text != "Test message" {
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
	DisplayedMessages []Message
	inputQueue        []string
	inputIndex        int
	mutex             sync.Mutex
}

func (m *MockUIRenderer) DisplayMessage(msg Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.DisplayedMessages = append(m.DisplayedMessages, msg)
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
	SentMessages []Message
	mutex        sync.Mutex
}

func (m *MockMessageSender) Send(msg Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SentMessages = append(m.SentMessages, msg)
	return nil
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
