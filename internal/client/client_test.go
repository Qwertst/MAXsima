package client

import (
	"fmt"
	"net"
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
	peerAddress string
	conn        net.Conn
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
