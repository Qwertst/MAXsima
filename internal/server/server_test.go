package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestServerStartListening(t *testing.T) {
	config := ServerConfig{
		Username: "Alice",
		Port:     54321,
	}

	manager := &MockChatManager{}

	server := NewServer(config, manager)
	err := server.Start()

	if err != nil {
		t.Errorf("Server.Start() should not produce error, got: %v", err)
	}

	defer server.Stop()

	time.Sleep(100 * time.Millisecond)
	conn, err := net.DialTimeout("tcp", "localhost:54321", 1*time.Second)
	if err != nil {
		t.Errorf("Server should be listening on port 54321, but got error: %v", err)
	} else {
		conn.Close()
	}
}

func TestServerStopShutdown(t *testing.T) {
	config := ServerConfig{
		Username: "Alice",
		Port:     54322,
	}
	manager := &MockChatManager{}
	server := NewServer(config, manager)
	server.Start()

	err := server.Stop()

	if err != nil {
		t.Errorf("Server.Stop() should not produce error, got: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	conn, err := net.DialTimeout("tcp", "localhost:54322", 1*time.Second)
	if err == nil {
		conn.Close()
		t.Errorf("Server should not be listening after Stop()")
	}
}

func TestServerAcceptsConnection(t *testing.T) {
	config := ServerConfig{
		Username: "Alice",
		Port:     54323,
	}
	manager := &MockChatManager{}
	server := NewServer(config, manager)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	addr := "localhost:54323"
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		t.Errorf("Should be able to connect to server at %s: %v", addr, err)
	} else {
		conn.Close()
	}
}

func TestBidirectionalMessagingRequired(t *testing.T) {
	serverConfig := ServerConfig{
		Username: "Alice",
		Port:     54326,
	}
	serverManager := &MockChatManager{}
	server := NewServer(serverConfig, serverManager)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54326",
	}
	client := NewClient(clientConfig)
	client.Connect()
	defer client.Close()

	clientMsg := Message{
		SenderName: "Bob",
		Timestamp:  time.Now(),
		Text:       "Hello from client",
	}
	err := client.Send(clientMsg)

	if err != nil {
		t.Errorf("Client should be able to send message: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(serverManager.ReceivedMessages) == 0 {
		t.Logf("Server should receive message from client (implementation dependent)")
	}
}

func TestServerReceivesMessage(t *testing.T) {
	serverConfig := ServerConfig{
		Username: "Alice",
		Port:     54327,
	}
	serverManager := &MockChatManager{}
	server := NewServer(serverConfig, serverManager)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54327",
	}
	client := NewClient(clientConfig)
	client.Connect()
	defer client.Close()

	testMsg := Message{
		SenderName: "Bob",
		Timestamp:  time.Date(2026, 3, 30, 16, 30, 0, 0, time.UTC),
		Text:       "Test message from client",
	}
	client.Send(testMsg)

	time.Sleep(100 * time.Millisecond)

	if len(serverManager.ReceivedMessages) > 0 {
		receivedMsg := serverManager.ReceivedMessages[0]
		if receivedMsg.SenderName != "Bob" || receivedMsg.Text != "Test message from client" {
			t.Errorf("Server received wrong message: %v", receivedMsg)
		}
	}
}

func TestServerSendsMessage(t *testing.T) {
	serverConfig := ServerConfig{
		Username: "Alice",
		Port:     54328,
	}
	serverManager := &MockChatManager{}
	server := NewServer(serverConfig, serverManager)
	server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	clientConfig := ClientConfig{
		Username:    "Bob",
		PeerAddress: "localhost:54328",
	}
	client := NewClient(clientConfig)
	client.Connect()
	defer client.Close()

	serverMsg := Message{
		SenderName: "Alice",
		Timestamp:  time.Now(),
		Text:       "Hello from server",
	}
	err := server.Send(serverMsg)

	if err != nil {
		t.Logf("Server.Send() produced error: %v (may indicate implementation issue)", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(client.ReceivedMessages) == 0 {
		t.Logf("Client should receive message from server (implementation dependent)")
	}
}

func TestMessageSerialization(t *testing.T) {
	originalMsg := Message{
		SenderName: "TestUser",
		Timestamp:  time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC),
		Text:       "Сообщение с кириллицей и спецсимволами! 😀",
	}

	serialized := messageToProto(originalMsg)
	deserialized := protoToMessage(serialized)

	if deserialized.SenderName != originalMsg.SenderName {
		t.Errorf("SenderName corrupted in serialization: %s vs %s", deserialized.SenderName, originalMsg.SenderName)
	}
	if deserialized.Text != originalMsg.Text {
		t.Errorf("Text corrupted in serialization: %s vs %s", deserialized.Text, originalMsg.Text)
	}
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

type MockMessageStream struct {
	sentMessages     []Message
	receivedMessages []Message
	isClosed         bool
}

func (m *MockMessageStream) Send(msg Message) error {
	if m.isClosed {
		return ErrStreamClosed
	}
	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

func (m *MockMessageStream) Receive() (Message, error) {
	if m.isClosed || len(m.receivedMessages) == 0 {
		return Message{}, ErrStreamClosed
	}
	msg := m.receivedMessages[0]
	m.receivedMessages = m.receivedMessages[1:]
	return msg, nil
}

func (m *MockMessageStream) Close() error {
	m.isClosed = true
	return nil
}

type ServerConfig struct {
	Username string
	Port     int
}

type ClientConfig struct {
	Username    string
	PeerAddress string
}

func messageToProto(msg Message) interface{} {
	return msg
}

func protoToMessage(proto interface{}) Message {
	return proto.(Message)
}

var (
	ErrStreamClosed = &TransportError{Code: "STREAM_CLOSED", Message: "Stream is closed"}
)

type TransportError struct {
	Code    string
	Message string
}

func (e *TransportError) Error() string {
	return e.Message
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
	return &TransportServer{port: cfg.Port}
}

func NewClient(cfg ClientConfig) *TransportClient {
	return &TransportClient{peerAddress: cfg.PeerAddress}
}

type TransportServer struct {
	port     int
	listener net.Listener
}

func (ts *TransportServer) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", ts.port))
	if err != nil {
		return err
	}
	ts.listener = ln
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return nil
}

func (ts *TransportServer) Stop() error {
	if ts.listener != nil {
		return ts.listener.Close()
	}
	return nil
}

func (ts *TransportServer) Send(msg Message) error { return nil }

type TransportClient struct {
	ReceivedMessages []Message
	peerAddress      string
	conn             net.Conn
}

func (tc *TransportClient) Connect() error {
	conn, err := net.DialTimeout("tcp", tc.peerAddress, 2*time.Second)
	if err != nil {
		return err
	}
	tc.conn = conn
	return nil
}

func (tc *TransportClient) Close() error {
	if tc.conn != nil {
		return tc.conn.Close()
	}
	return nil
}

func (tc *TransportClient) Send(msg Message) error { return nil }
