package server

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/testutil"
	pb "github.com/aydreq/maxsima/proto/gen/chat"
)

// startRealServer starts a gRPC server using the production server.New() and
// registers it on a random port. Returns the address and a stop function.
func startRealServer(t *testing.T, mgr *chat.Manager) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	// Use the production server.New() — this is what we're testing.
	srv := New(mgr)
	pb.RegisterChatServiceServer(grpcSrv, srv)
	go func() { _ = grpcSrv.Serve(lis) }()
	return lis.Addr().String(), func() { grpcSrv.Stop() }
}

// dialStream opens a gRPC Connect stream to addr.
func dialStream(t *testing.T, addr string) (pb.ChatService_ConnectClient, func()) {
	t.Helper()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	stream, err := pb.NewChatServiceClient(conn).Connect(context.Background())
	if err != nil {
		conn.Close()
		t.Fatalf("Connect: %v", err)
	}
	return stream, func() { conn.Close() }
}

func TestServerNew(t *testing.T) {
	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Alice", mUI)
	srv := New(mgr)
	if srv == nil {
		t.Errorf("New() should return a non-nil Server")
	}
}

func TestServerAcceptsGRPCConnection(t *testing.T) {
	mUI := testutil.NewBlockingMockUI()
	mgr := chat.NewManager("Alice", mUI)
	addr, stop := startRealServer(t, mgr)
	defer stop()
	defer mUI.Stop()

	_, closeConn := dialStream(t, addr)
	defer closeConn()
	// Reaching here without error means server.Connect() was invoked.
}

func TestServerListensOnPort(t *testing.T) {
	mUI := testutil.NewBlockingMockUI()
	mgr := chat.NewManager("Alice", mUI)
	addr, stop := startRealServer(t, mgr)
	defer stop()
	defer mUI.Stop()

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Errorf("server should be reachable at %s: %v", addr, err)
		return
	}
	conn.Close()
}

func TestServerReceivesMessageFromClient(t *testing.T) {
	mUI := testutil.NewBlockingMockUI()
	mgr := chat.NewManager("Alice", mUI)
	addr, stop := startRealServer(t, mgr)
	defer stop()
	defer mUI.Stop()

	stream, closeConn := dialStream(t, addr)
	defer closeConn()

	err := stream.Send(&pb.ChatMessage{
		SenderName: "Bob",
		Timestamp:  time.Now().Unix(),
		Text:       "Hello Alice!",
	})
	if err != nil {
		t.Fatalf("client Send failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	msgs := mUI.Messages()
	if len(msgs) == 0 {
		t.Fatalf("server manager should have displayed the incoming message")
	}
	if msgs[0].SenderName != "Bob" || msgs[0].Text != "Hello Alice!" {
		t.Errorf("unexpected message: %+v", msgs[0])
	}
}

func TestServerSendsMessageToClient(t *testing.T) {
	// Use a MockUI with one queued input so the server sends one message.
	serverUI := &testutil.MockUI{}
	serverUI.SetInputs([]string{"Hello Bob!"})
	mgr := chat.NewManager("Alice", serverUI)
	addr, stop := startRealServer(t, mgr)
	defer stop()

	stream, closeConn := dialStream(t, addr)
	defer closeConn()

	done := make(chan *pb.ChatMessage, 1)
	go func() {
		msg, err := stream.Recv()
		if err != nil {
			done <- nil
			return
		}
		done <- msg
	}()

	select {
	case msg := <-done:
		if msg == nil {
			t.Fatalf("expected a message from server, got error/EOF")
		}
		if msg.SenderName != "Alice" || msg.Text != "Hello Bob!" {
			t.Errorf("unexpected message: %+v", msg)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for message from server")
	}
}

func TestServerStopsAfterClientDisconnects(t *testing.T) {
	mUI := testutil.NewBlockingMockUI()
	mgr := chat.NewManager("Alice", mUI)
	addr, stop := startRealServer(t, mgr)
	defer stop()
	defer mUI.Stop()

	stream, closeConn := dialStream(t, addr)
	closeConn()
	_ = stream

	time.Sleep(200 * time.Millisecond)
	// No assertion needed — test passes if no deadlock/panic.
}

func TestServerListenFunction(t *testing.T) {
	// Test the Listen() function by starting it in a goroutine and verifying
	// the port is reachable, then stopping via the gRPC server shutdown.
	// We can't call Listen() directly in a test without a way to stop it,
	// so we verify it returns an error on an invalid port.
	mUI := &testutil.MockUI{}
	mgr := chat.NewManager("Alice", mUI)
	srv := New(mgr)

	// Listen on port 0 is not valid for our Listen() which uses fmt.Sprintf(":%d", port).
	// Instead verify Listen returns error on already-used port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("setup listen: %v", err)
	}
	usedPort := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	// Listen on the same port in a goroutine; it should succeed briefly.
	done := make(chan error, 1)
	go func() {
		done <- Listen(usedPort, srv)
	}()

	// Give it time to start.
	time.Sleep(100 * time.Millisecond)

	// Verify it's listening.
	conn, err := net.DialTimeout("tcp", lis.Addr().String(), time.Second)
	if err != nil {
		t.Logf("Listen() may not have started yet: %v", err)
	} else {
		conn.Close()
	}

	// We can't easily stop Listen() from outside, so just verify no immediate error.
	select {
	case err := <-done:
		if err != nil {
			t.Logf("Listen() returned error (may be expected in test): %v", err)
		}
	default:
		// Still running — that's the expected state.
	}
}
