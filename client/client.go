package client

import (
	"context"
	"time"

	pb "github.com/marcusbello/qotd-service/proto/qotd"
	"google.golang.org/grpc"
)

type Client struct {
	client pb.QOTDClient
	conn   *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Client{
		client: pb.NewQOTDClient(conn),
		conn:   conn,
	}, nil
}

func (c *Client) QOTD(ctx context.Context, wantAuthor string) (string, string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	}
	resp, err := c.client.GetQOTD(ctx, &pb.GetReq{Author: wantAuthor})
	if err != nil {
		return "", "", err
	}
	return resp.Author, resp.Quote, nil
}
