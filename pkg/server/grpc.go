package server

import (
	"fmt"
	"net"
	"tuber/pkg/proto"
	"tuber/pkg/reviewapps"

	"google.golang.org/grpc"
)

// Start starts a GRPC server
func Start() error {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}

	s := reviewapps.Server{}

	server := grpc.NewServer()
	proto.RegisterTuberServer(server, &s)

	fmt.Println("starting GRPC server")
	if err := server.Serve(lis); err != nil {
		return err
	}

	return nil
}
