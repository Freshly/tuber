package server

import (
	"fmt"
	"net"
	"tuber/pkg/proto"
	"tuber/pkg/reviewapps"

	"google.golang.org/grpc"
)

// Start starts a GRPC server
func Start(port int, s reviewapps.Server) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	proto.RegisterTuberServer(server, &s)

	fmt.Println("starting GRPC server")
	if err := server.Serve(lis); err != nil {
		return err
	}

	return nil
}
