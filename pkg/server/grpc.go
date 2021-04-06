package server

import (
	"fmt"
	"net"

	"github.com/freshly/tuber/pkg/proto"
	"github.com/freshly/tuber/pkg/reviewapps"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Start starts a GRPC server
func Start(port int, s reviewapps.Server) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	proto.RegisterTuberServer(server, &s)
	reflection.Register(server)

	if err := server.Serve(lis); err != nil {
		return err
	}

	return nil
}
