package client

import (
	"log"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
)

func NewClient() proto.TuberServiceClient {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	client := proto.NewTuberServiceClient(conn)

	return client
}
