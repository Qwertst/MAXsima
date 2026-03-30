package client

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/testutil"
	pb "github.com/aydreq/maxsima/proto/gen/chat"
)

// echoServer is a minimal gRPC ChatService that echoes every received message
// back to the sender.
type echoServer struct {
	pb.UnimplementedChatServiceServer
}

func (e *echoServer) Connect(stream pb.ChatService_ConnectServer) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}

// noopServer accepts connections but never sends or receives anything.
type noopServer struct {
	pb.UnimplementedChatServiceServer
	connected chan struct{}
}

func (n *noopServer) Connect(stream pb.ChatService_ConnectServer) error {
	close(n.connected)
	<-stream.Context().Done()
	return nil
}

// startTestServer starts a gRPC server with the given handler on a random port
// and returns the address and a stop function.
func startTestServer(t *testing.T, srv pb.ChatServiceServer) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcSrv, srv)
	go func() { _ = grpcSrv.Serve(lis) }()
	return lis.Addr().String(), func() { grpcSrv.Stop() }
}

func TestClientConnectsToServer(t *testing.T) {
	ns := &noopServer{connected: make(chan struct{})}
	addr, stop := startTestServer(t, ns)
	defer stop()

	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Bob", mUI)

	c, err := New(addr, mgr)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer c.conn.Close()

	select {
	case <-ns.connected:
		// server saw the connection — pass
	case <-time.After(2 * time.Second):
		t.Errorf("server did not receive connection within timeout")
	}
}

func TestClientConnectFailsOnNoServer(t *testing.T) {
	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Bob", mUI)
	// Port 1 is never open in test environments.
	_, err := New("127.0.0.1:1", mgr)
	if err == nil {
		t.Errorf("New() should fail when no server is listening")
	}
}

func TestClientNewFailsWhenServerRejectsGRPC(t *testing.T) {
	// Start a plain TCP server that immediately closes connections.
	// This causes grpc.Dial to succeed (TCP connects) but stub.Connect to fail.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				return
			}
			conn.Close() // immediately close — not a gRPC server
		}
	}()
	defer lis.Close()

	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Bob", mUI)
	_, err = New(lis.Addr().String(), mgr)
	if err == nil {
		t.Errorf("New() should fail when server is not a gRPC server")
	}
}

func TestClientRunSendsAndReceivesMessages(t *testing.T) {
	// echoServer echoes messages back. We send one message; the UI then
	// returns io.EOF which stops the outgoing handler. The incoming handler
	// should display the echoed message before the session fully closes.
	// We poll for the displayed message with a generous timeout.
	addr, stop := startTestServer(t, &echoServer{})
	defer stop()

	mUI := &testutil.MockUI{}
	mUI.SetInputs([]string{"Hello server!"})
	mgr := chat.NewManager("Bob", mUI)

	c, err := New(addr, mgr)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	go c.Run() //nolint:errcheck

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		msgs := mUI.Messages()
		if len(msgs) > 0 {
			if msgs[0].Text != "Hello server!" {
				t.Errorf("unexpected displayed message: %+v", msgs[0])
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("timed out waiting for echoed message to be displayed")
}

func TestClientHandlesServerDisconnect(t *testing.T) {
	ns := &noopServer{connected: make(chan struct{})}
	addr, stop := startTestServer(t, ns)

	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Bob", mUI)

	c, err := New(addr, mgr)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- c.Run() }()

	// Wait for the server to see the connection, then shut it down.
	<-ns.connected
	stop()

	select {
	case <-done:
		// Run() returned after server disconnect — pass
	case <-time.After(3 * time.Second):
		t.Errorf("Run() did not return after server disconnect")
	}
}
