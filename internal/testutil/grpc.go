package testutil

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/model"
	pb "github.com/aydreq/maxsima/proto/gen/chat"
)

// StartGRPCServer starts a real gRPC server backed by a chat.Manager on a
// random port. The server's UI is a BlockingMockUI so the session stays alive
// until the returned stop function is called (which also unblocks the UI).
// Returns the address, the server's BlockingMockUI (for asserting received
// messages), and a stop function.
func StartGRPCServer(t *testing.T, username string) (addr string, serverUI *BlockingMockUI, stop func()) {
	t.Helper()
	serverUI = NewBlockingMockUI()
	mgr := chat.NewManager(username, serverUI)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testutil.StartGRPCServer: listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcSrv, &serverAdapter{mgr: mgr})
	go func() { _ = grpcSrv.Serve(lis) }()

	return lis.Addr().String(), serverUI, func() {
		serverUI.Stop() // unblock ReadInput so the session ends cleanly
		grpcSrv.Stop()
	}
}

// StartGRPCServerWithManager starts a gRPC server using the provided manager.
// The caller is responsible for ensuring the manager's UI does not return
// io.EOF prematurely (use BlockingMockUI for the server side).
func StartGRPCServerWithManager(t *testing.T, mgr *chat.Manager) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testutil.StartGRPCServerWithManager: listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcSrv, &serverAdapter{mgr: mgr})
	go func() { _ = grpcSrv.Serve(lis) }()
	return lis.Addr().String(), func() { grpcSrv.Stop() }
}

// DialGRPC dials addr, opens a Connect stream, starts a chat.Manager session
// backed by a fresh MockUI, and returns the manager, its MockUI, and a close
// function.
func DialGRPC(t *testing.T, addr, username string) (mgr *chat.Manager, mUI *MockUI, closeFn func()) {
	t.Helper()
	mUI = &MockUI{}
	mgr = chat.NewManager(username, mUI)
	closeFn = dialWithManager(t, addr, mgr)
	return mgr, mUI, closeFn
}

// DialGRPCWithManager dials addr and starts a session using the provided
// manager. Callers hold their own UI reference. Returns a close function.
func DialGRPCWithManager(t *testing.T, addr string, mgr *chat.Manager) (*chat.Manager, *MockUI, func()) {
	t.Helper()
	closeFn := dialWithManager(t, addr, mgr)
	return mgr, nil, closeFn
}

// TryDialGRPC attempts to dial addr and open a Connect stream without failing
// the test. Returns an error if the connection or stream could not be opened.
func TryDialGRPC(addr, username string) (*chat.Manager, func(), error) {
	mUI := &MockUI{}
	mgr := chat.NewManager(username, mUI)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	stream, err := pb.NewChatServiceClient(conn).Connect(context.Background())
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	ca := &clientStreamAdapter{stream: stream}
	if err := mgr.StartSession(ca, ca); err != nil {
		conn.Close()
		return nil, nil, err
	}
	return mgr, func() { _ = mgr.StopSession(); conn.Close() }, nil
}

// dialWithManager is the shared implementation for DialGRPC / DialGRPCWithManager.
func dialWithManager(t *testing.T, addr string, mgr *chat.Manager) func() {
	t.Helper()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("testutil.dialWithManager: dial: %v", err)
	}
	stream, err := pb.NewChatServiceClient(conn).Connect(context.Background())
	if err != nil {
		conn.Close()
		t.Fatalf("testutil.dialWithManager: Connect: %v", err)
	}
	ca := &clientStreamAdapter{stream: stream}
	if err := mgr.StartSession(ca, ca); err != nil {
		conn.Close()
		t.Fatalf("testutil.dialWithManager: StartSession: %v", err)
	}
	return func() {
		_ = mgr.StopSession()
		conn.Close()
	}
}

// ---------------------------------------------------------------------------
// serverAdapter — wires pb.ChatServiceServer to chat.Manager.
// Mirrors internal/server.Server without importing it (avoids import cycle).
// ---------------------------------------------------------------------------

type serverAdapter struct {
	pb.UnimplementedChatServiceServer
	mgr *chat.Manager
}

func (s *serverAdapter) Connect(stream pb.ChatService_ConnectServer) error {
	a := &serverStreamAdapter{stream: stream}
	if err := s.mgr.StartSession(a, a); err != nil {
		return err
	}
	s.mgr.Wait()
	return nil
}

type serverStreamAdapter struct {
	stream pb.ChatService_ConnectServer
}

func (a *serverStreamAdapter) Send(msg model.Message) error {
	return a.stream.Send(&pb.ChatMessage{
		SenderName: msg.SenderName,
		Timestamp:  msg.Timestamp.Unix(),
		Text:       msg.Text,
	})
}

func (a *serverStreamAdapter) Receive() (model.Message, error) {
	pbMsg, err := a.stream.Recv()
	if err != nil {
		return model.Message{}, err
	}
	return model.Message{
		SenderName: pbMsg.SenderName,
		Timestamp:  time.Unix(pbMsg.Timestamp, 0),
		Text:       pbMsg.Text,
	}, nil
}

// ---------------------------------------------------------------------------
// clientStreamAdapter — wires pb.ChatService_ConnectClient to chat session.
// ---------------------------------------------------------------------------

type clientStreamAdapter struct {
	stream pb.ChatService_ConnectClient
}

func (a *clientStreamAdapter) Send(msg model.Message) error {
	return a.stream.Send(&pb.ChatMessage{
		SenderName: msg.SenderName,
		Timestamp:  msg.Timestamp.Unix(),
		Text:       msg.Text,
	})
}

func (a *clientStreamAdapter) Receive() (model.Message, error) {
	pbMsg, err := a.stream.Recv()
	if err != nil {
		return model.Message{}, err
	}
	return model.Message{
		SenderName: pbMsg.SenderName,
		Timestamp:  time.Unix(pbMsg.Timestamp, 0),
		Text:       pbMsg.Text,
	}, nil
}
