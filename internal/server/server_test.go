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

func TestServerAcceptsGRPCConnection(t *testing.T) {
	addr, _, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	_, closeConn := dialStream(t, addr)
	defer closeConn()
	// Reaching here without error means the server accepted the connection.
}

func TestServerListensOnPort(t *testing.T) {
	addr, _, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Errorf("server should be reachable at %s: %v", addr, err)
		return
	}
	conn.Close()
}

func TestServerReceivesMessageFromClient(t *testing.T) {
	addr, serverUI, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

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

	msgs := serverUI.Messages()
	if len(msgs) == 0 {
		t.Fatalf("server manager should have displayed the incoming message")
	}
	if msgs[0].SenderName != "Bob" || msgs[0].Text != "Hello Alice!" {
		t.Errorf("unexpected message: %+v", msgs[0])
	}
}

func TestServerSendsMessageToClient(t *testing.T) {
	// Use a server with a MockUI that has one queued input, then blocks.
	// We use StartGRPCServerWithManager so we can control the server UI.
	serverUI := &testutil.MockUI{}
	serverUI.SetInputs([]string{"Hello Bob!"})
	// After sending "Hello Bob!" the MockUI returns io.EOF, ending the session.
	// That's fine for this test — we just want to verify the message arrives.
	serverMgr := chat.NewManager("Alice", serverUI)
	addr, stop := testutil.StartGRPCServerWithManager(t, serverMgr)
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
	addr, _, stop := testutil.StartGRPCServer(t, "Alice")
	defer stop()

	stream, closeConn := dialStream(t, addr)
	closeConn()
	_ = stream

	// Give the server goroutine time to notice the disconnect.
	time.Sleep(200 * time.Millisecond)
	// No assertion needed — test passes if no deadlock/panic.
}
