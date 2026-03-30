package integration

import (
	"testing"
	"time"
)

func TestFullServerModeInitialization(t *testing.T) {
	cliArgs := []string{
		"--username", "Alice",
		"--port", "54330",
	}

	app := InitializeApplication(cliArgs)
	err := app.Start()

	if err != nil {
		t.Errorf("Server mode initialization failed: %v", err)
	}

	defer app.Stop()

	time.Sleep(100 * time.Millisecond)
	if !app.IsListening() {
		t.Errorf("Server should be listening after Start()")
	}
}

func TestFullClientModeInitialization(t *testing.T) {
	serverApp := InitializeApplication([]string{
		"--username", "Alice",
		"--port", "54331",
	})
	serverApp.Start()
	defer serverApp.Stop()

	time.Sleep(100 * time.Millisecond)

	cliArgs := []string{
		"--username", "Bob",
		"--peer", "localhost:54331",
	}

	clientApp := InitializeApplication(cliArgs)
	err := clientApp.Start()

	if err != nil {
		t.Errorf("Client mode initialization failed: %v", err)
	}

	defer clientApp.Stop()

	time.Sleep(100 * time.Millisecond)
	if !clientApp.IsConnected() {
		t.Errorf("Client should be connected after Start()")
	}
}

func TestPeerToPeerMessaging(t *testing.T) {
	serverApp := InitializeApplication([]string{
		"--username", "Alice",
		"--port", "54332",
	})
	serverApp.Start()
	defer serverApp.Stop()

	time.Sleep(100 * time.Millisecond)

	clientApp := InitializeApplication([]string{
		"--username", "Bob",
		"--peer", "localhost:54332",
	})
	clientApp.Start()
	defer clientApp.Stop()

	time.Sleep(100 * time.Millisecond)

	clientApp.SendMessage("Hello Alice!")

	time.Sleep(100 * time.Millisecond)
	serverApp.SendMessage("Hi Bob!")

	time.Sleep(100 * time.Millisecond)

	bobMessages := clientApp.GetDisplayedMessages()
	aliceMessage := findMessageWithText(bobMessages, "Hi Bob!")
	if aliceMessage == nil {
		t.Errorf("Bob should receive message from Alice")
	} else {
		if aliceMessage.SenderName != "Alice" {
			t.Errorf("Message should show Alice as sender")
		}
		if aliceMessage.Timestamp.IsZero() {
			t.Errorf("Message should have timestamp")
		}
	}

	aliceMessages := serverApp.GetDisplayedMessages()
	bobMessage := findMessageWithText(aliceMessages, "Hello Alice!")
	if bobMessage == nil {
		t.Errorf("Alice should receive message from Bob")
	} else {
		if bobMessage.SenderName != "Bob" {
			t.Errorf("Message should show Bob as sender")
		}
	}
}

func TestConfigValidationOnStartup(t *testing.T) {
	invalidArgs := []string{
		"--port", "54333",
	}

	app := InitializeApplication(invalidArgs)
	err := app.Start()

	if err == nil {
		t.Errorf("Application should fail with invalid config (missing --username)")
	}
}

func TestModeDetectionFromCLI(t *testing.T) {
	serverArgs := []string{
		"--username", "Alice",
		"--port", "54334",
	}

	serverConfig := ParseCLI(serverArgs)

	if !serverConfig.IsServerMode() {
		t.Errorf("Config should be detected as server mode when --peer is absent")
	}

	clientArgs := []string{
		"--username", "Bob",
		"--peer", "localhost:54334",
	}

	clientConfig := ParseCLI(clientArgs)

	if clientConfig.IsServerMode() {
		t.Errorf("Config should be detected as client mode when --peer is present")
	}
}

func TestGracefulShutdown(t *testing.T) {
	cliArgs := []string{
		"--username", "Alice",
		"--port", "54335",
	}
	app := InitializeApplication(cliArgs)
	app.Start()

	time.Sleep(100 * time.Millisecond)

	err := app.Stop()

	if err != nil {
		t.Errorf("Graceful shutdown should not produce error, got: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if app.IsRunning() {
		t.Errorf("Application should not be running after Stop()")
	}
}

func TestConnectionInterruption(t *testing.T) {
	serverApp := InitializeApplication([]string{
		"--username", "Alice",
		"--port", "54336",
	})
	serverApp.Start()
	defer serverApp.Stop()

	time.Sleep(100 * time.Millisecond)

	clientApp := InitializeApplication([]string{
		"--username", "Bob",
		"--peer", "localhost:54336",
	})
	clientApp.Start()
	defer clientApp.Stop()

	time.Sleep(100 * time.Millisecond)

	serverApp.Stop()

	time.Sleep(100 * time.Millisecond)

	if clientApp.IsPanickedDueToConnectionError() {
		t.Errorf("Client should handle connection interruption gracefully")
	}
}

func TestConcurrentBidirectionalMessaging(t *testing.T) {
	serverApp := InitializeApplication([]string{
		"--username", "Alice",
		"--port", "54337",
	})
	serverApp.Start()
	defer serverApp.Stop()

	time.Sleep(100 * time.Millisecond)

	clientApp := InitializeApplication([]string{
		"--username", "Bob",
		"--peer", "localhost:54337",
	})
	clientApp.Start()
	defer clientApp.Stop()

	time.Sleep(100 * time.Millisecond)

	go serverApp.SendMessage("Message 1 from Alice")
	go clientApp.SendMessage("Message 1 from Bob")

	go serverApp.SendMessage("Message 2 from Alice")
	go clientApp.SendMessage("Message 2 from Bob")

	time.Sleep(200 * time.Millisecond)

	bobMessages := clientApp.GetDisplayedMessages()
	if countMessagesFromSender(bobMessages, "Alice") < 2 {
		t.Errorf("Bob should receive both messages from Alice (concurrent messaging)")
	}

	aliceMessages := serverApp.GetDisplayedMessages()
	if countMessagesFromSender(aliceMessages, "Bob") < 2 {
		t.Errorf("Alice should receive both messages from Bob (concurrent messaging)")
	}
}

type Application struct {
	displayedMessages []Message
	isListening       bool
	isConnected       bool
	isRunning         bool
	panickedFlag      bool
}

type Message struct {
	SenderName string
	Timestamp  time.Time
	Text       string
}

func InitializeApplication(cliArgs []string) *Application {
	return &Application{
		displayedMessages: []Message{},
		isRunning:         false,
	}
}

func (app *Application) Start() error {
	app.isRunning = true
	app.isListening = true
	app.isConnected = true
	return nil
}

func (app *Application) Stop() error {
	app.isRunning = false
	app.isListening = false
	app.isConnected = false
	return nil
}

func (app *Application) SendMessage(text string) {
}

func (app *Application) GetDisplayedMessages() []Message {
	return app.displayedMessages
}

func (app *Application) IsListening() bool {
	return app.isListening
}

func (app *Application) IsConnected() bool {
	return app.isConnected
}

func (app *Application) IsRunning() bool {
	return app.isRunning
}

func (app *Application) IsPanickedDueToConnectionError() bool {
	return app.panickedFlag
}

func ParseCLI(args []string) *Config {
	return &Config{}
}

type Config struct {
	isServer bool
}

func (c *Config) IsServerMode() bool {
	return c.isServer
}

func findMessageWithText(messages []Message, text string) *Message {
	for i := range messages {
		if messages[i].Text == text {
			return &messages[i]
		}
	}
	return nil
}

func countMessagesFromSender(messages []Message, sender string) int {
	count := 0
	for i := range messages {
		if messages[i].SenderName == sender {
			count++
		}
	}
	return count
}
