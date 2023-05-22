package server

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"

	pb "github.com/marcusbello/qotd-service/proto/qotd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	pb.UnimplementedQOTDServer
	addr       string
	quotes     map[string][]string
	mu         sync.Mutex
	grpcServer *grpc.Server
}

func New(addr string) (*API, error) {
	var opts []grpc.ServerOption
	a := &API{
		addr: addr,
		quotes: map[string][]string{
			// insert list of quotes
			"Kelvin Hart": {
				"My son's password is F.U",
				"When our kids caught me and my wife",
			},
			"Dave Chapelle": {
				"Always in trouble with the commmunity",
				"Another day to attack the community",
			},
		},
		grpcServer: grpc.NewServer(opts...),
	}
	a.grpcServer.RegisterService(&pb.QOTD_ServiceDesc, a)
	return a, nil
}

func (a *API) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	lis, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}

	return a.grpcServer.Serve(lis)
}

func (a *API) GetQOTD(ctx context.Context, req *pb.GetReq) (*pb.GetResp, error) {
	var (
		author string
		quotes []string
	)

	if req.Author == "" {
		for author, quotes = range a.quotes {
			break
		}
	} else {
		author = req.Author
		var ok bool
		quotes, ok = a.quotes[req.Author]
		if !ok {
			return nil, status.Error(
				codes.NotFound,
				fmt.Sprintf("author %q not found", req.Author),
			)
		}
	}

	return &pb.GetResp{
		Author: author,
		Quote:  quotes[rand.Intn(len(quotes))],
	}, nil
}
