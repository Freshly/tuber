package client

import (
	"fmt"
	"log"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
)

// NewClient returns a GRPC client
func NewClient(url string) (proto.TuberServiceClient, *grpc.ClientConn) {
	fullURL := fmt.Sprintf("%s:9000", url)

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fullURL, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}

	client := proto.NewTuberServiceClient(conn)

	return client, conn
}
