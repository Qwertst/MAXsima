package server

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/model"
	pb "github.com/aydreq/maxsima/proto/gen/chat"
)

type Server struct {
	pb.UnimplementedChatServiceServer
	manager *chat.Manager
}

func New(manager *chat.Manager) *Server {
	return &Server{manager: manager}
}

func (s *Server) Connect(stream pb.ChatService_ConnectServer) error {
	adapter := &streamAdapter{stream: stream}
	if err := s.manager.StartSession(adapter, adapter); err != nil {
		return err
	}
	s.manager.Wait()
	return nil
}

func Listen(port int, srv *Server) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterChatServiceServer(grpcSrv, srv)
	fmt.Printf("Listening on :%d...\n", port)
	return grpcSrv.Serve(lis)
}

type streamAdapter struct {
	stream pb.ChatService_ConnectServer
}

func (a *streamAdapter) Send(msg model.Message) error {
	return a.stream.Send(&pb.ChatMessage{
		SenderName: msg.SenderName,
		Timestamp:  msg.Timestamp.Unix(),
		Text:       msg.Text,
	})
}

func (a *streamAdapter) Receive() (model.Message, error) {
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
