package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aydreq/maxsima/config"
	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/testutil"
)

// ---------------------------------------------------------------------------
// Config tests
// ---------------------------------------------------------------------------

func TestConfigServerMode(t *testing.T) {
	cfg := config.ParseFlags([]string{"--username", "Alice", "--port", "54340"})
	if !cfg.IsServerMode() {
		t.Errorf("expected server mode when --peer is absent")
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid server config should not fail Validate: %v", err)
	}
}

func TestConfigClientMode(t *testing.T) {
	cfg := config.ParseFlags([]string{"--username", "Bob", "--peer", "localhost:54340"})
	if cfg.IsServerMode() {
		t.Errorf("expected client mode when --peer is present")
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid client config should not fail Validate: %v", err)
	}
}

func TestConfigValidationMissingUsername(t *testing.T) {
	cfg := config.ParseFlags([]string{"--port", "54341"})
	if err := cfg.Validate(); err == nil {
		t.Errorf("config without --username should fail Validate")
	}
}

func TestValidatePeerAddress(t *testing.T) {
	cases := []struct {
		addr    string
		wantErr bool
	}{
		{"localhost:8080", false},
		{"127.0.0.1:54321", false},
		{"", true},
		{"noport", true},
		{"host:notaport", true},
		{fmt.Sprintf("host:%d", 0), true},
	}
	for _, tc := range cases {
		cfg := &config.Config{PeerAddress: tc.addr}
		err := cfg.ValidatePeerAddress()
		if tc.wantErr && err == nil {
			t.Errorf("ValidatePeerAddress(%q) expected error, got nil", tc.addr)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidatePeerAddress(%q) unexpected error: %v", tc.addr, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Server / client integration tests
// ---------------------------------------------------------------------------

func TestServerListens(t *testing.T) {
	addr, _, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	_, _, closeClient := testutil.DialGRPC(t, addr, "Bob")
	defer closeClient()
	// Reaching here without error means the server is reachable.
}

func TestServerReceivesMessageFromClient(t *testing.T) {
	addr, serverUI, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	clientUI := &testutil.MockUI{}
	clientUI.SetInputs([]string{"Hello Alice!"})
	clientMgr := chat.NewManager("Bob", clientUI)
	_, _, closeClient := testutil.DialGRPCWithManager(t, addr, clientMgr)
	defer closeClient()

	time.Sleep(300 * time.Millisecond)

	msgs := serverUI.Messages()
	found := false
	for _, m := range msgs {
		if m.Text == "Hello Alice!" && m.SenderName == "Bob" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("server did not display expected message; got: %v", msgs)
	}
}

func TestClientReceivesMessageFromServer(t *testing.T) {
	// Use a server with a MockUI that sends one message then EOF.
	serverUI := &testutil.MockUI{}
	serverUI.SetInputs([]string{"Hello Bob!"})
	serverMgr := chat.NewManager("Alice", serverUI)
	addr, stop := testutil.StartGRPCServerWithManager(t, serverMgr)
	defer stop()

	clientUI := &testutil.MockUI{}
	clientMgr := chat.NewManager("Bob", clientUI)
	_, _, closeClient := testutil.DialGRPCWithManager(t, addr, clientMgr)
	defer closeClient()

	// Poll for the message with a generous timeout.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		msgs := clientUI.Messages()
		for _, m := range msgs {
			if m.Text == "Hello Bob!" && m.SenderName == "Alice" {
				return // pass
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("client did not display expected message from server; got: %v", clientUI.Messages())
}

func TestGracefulShutdown(t *testing.T) {
	addr, _, stop := testutil.StartGRPCServer(t, "Alice")

	_, _, closeClient := testutil.DialGRPC(t, addr, "Bob")
	closeClient()

	stop()

	time.Sleep(100 * time.Millisecond)

	// After stop, new connections should fail.
	_, _, err := testutil.TryDialGRPC(addr, "Charlie")
	if err == nil {
		t.Errorf("server should not accept connections after stop")
	}
}

func TestConcurrentMessages(t *testing.T) {
	addr, serverUI, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	clientUI := &testutil.MockUI{}
	clientUI.SetInputs([]string{"msg1", "msg2", "msg3"})
	clientMgr := chat.NewManager("Bob", clientUI)
	_, _, closeClient := testutil.DialGRPCWithManager(t, addr, clientMgr)
	defer closeClient()

	// Poll until all 3 messages arrive or timeout.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(serverUI.Messages()) >= 3 {
			return // pass
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("server should have received 3 messages, got %d", len(serverUI.Messages()))
}
