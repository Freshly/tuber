package reviewapps

import (
	"log"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
)

// NewClient returns a GRPC client
func NewClient(url string) (proto.TuberClient, *grpc.ClientConn) {
	// fullURL := fmt.Sprintf("%s:9000", url)

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9000", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("grpc client: %s", err)
	}

	return proto.NewTuberClient(conn), conn
}
