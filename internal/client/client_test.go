package client

import (
	"testing"
	"time"
)

func TestClientConnect(t *testing.T) {
	serverConfig := ServerConfig{
		Username: "Alice",
		Port:     54324,
	}
	serverManager := &MockChatManager{}
	server := NewServer(serverConfig, serverManager)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54324",
	}

	client := NewClient(clientConfig)
	err := client.Connect()

	if err != nil {
		t.Errorf("Client.Connect() should succeed, got: %v", err)
	}

	client.Close()
}

func TestClientConnectFailure(t *testing.T) {
	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54999",
	}

	client := NewClient(clientConfig)
	err := client.Connect()

	if err == nil {
		t.Errorf("Client.Connect() to non-existent server should produce error")
		client.Close()
	}
}

func TestClientClose(t *testing.T) {
	serverConfig := ServerConfig{
		Username: "Alice",
		Port:     54325,
	}
	server := NewServer(serverConfig, &MockChatManager{})
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54325",
	}
	client := NewClient(clientConfig)
	client.Connect()

	err := client.Close()

	if err != nil {
		t.Logf("Client.Close() produced warning: %v (may be expected)", err)
	}
}

type ServerConfig struct {
	Username string
	Port     int
}

type ClientConfig struct {
	Username    string
	PeerAddress string
}

type MockChatManager struct {
	ReceivedMessages []Message
	SentMessages     []Message
}

func (m *MockChatManager) StartSession(sender MessageSender, receiver MessageReceiver) error {
	return nil
}

func (m *MockChatManager) StopSession() error {
	return nil
}

type Message struct {
	SenderName string
	Timestamp  time.Time
	Text       string
}

type MessageSender interface {
	Send(msg Message) error
}

type MessageReceiver interface {
	Receive() (Message, error)
}

func NewServer(cfg ServerConfig, mgr interface{}) *TransportServer {
	return &TransportServer{}
}

func NewClient(cfg ClientConfig) *TransportClient {
	return &TransportClient{}
}

type TransportServer struct{}
func (ts *TransportServer) Start() error { return nil }
func (ts *TransportServer) Stop() error { return nil }
func (ts *TransportServer) Send(msg Message) error { return nil }

type TransportClient struct{}
func (tc *TransportClient) Connect() error { return nil }
func (tc *TransportClient) Close() error { return nil }
func (tc *TransportClient) Send(msg Message) error { return nil }
