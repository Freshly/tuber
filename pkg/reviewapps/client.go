package reviewapps

import (
	"fmt"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
)

// NewClient returns a GRPC client
func NewClient(url string) (proto.TuberClient, *grpc.ClientConn, error) {
	hostname := fmt.Sprintf("%s:9000", url)

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(hostname, grpc.WithInsecure())

	if err != nil {
		return nil, nil, fmt.Errorf("grpc client: %s", err)
	}

	return proto.NewTuberClient(conn), conn, nil
}
