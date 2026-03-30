package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/model"
	pb "github.com/aydreq/maxsima/proto/gen/chat"
)

type Client struct {
	conn    *grpc.ClientConn
	stream  pb.ChatService_ConnectClient
	manager *chat.Manager
}

func New(peerAddress string, manager *chat.Manager) (*Client, error) {
	conn, err := grpc.Dial(peerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	stub := pb.NewChatServiceClient(conn)
	stream, err := stub.Connect(context.Background())
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Client{conn: conn, stream: stream, manager: manager}, nil
}

func (c *Client) Run() error {
	defer c.conn.Close()
	adapter := &streamAdapter{stream: c.stream}
	if err := c.manager.StartSession(adapter, adapter); err != nil {
		return err
	}
	c.manager.Wait()
	return nil
}

type streamAdapter struct {
	stream pb.ChatService_ConnectClient
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
