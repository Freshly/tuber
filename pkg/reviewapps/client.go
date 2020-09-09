package reviewapps

import (
	"crypto/tls"
	"fmt"
	"tuber/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewClient returns a GRPC client
func NewClient(url string) (proto.TuberClient, *grpc.ClientConn, error) {
	hostname := fmt.Sprintf("%s:9000", url)

	var conn *grpc.ClientConn
	creds := credentials.NewTLS(&tls.Config{})

	conn, err := grpc.Dial(hostname, grpc.WithTransportCredentials(creds))

	if err != nil {
		return nil, nil, fmt.Errorf("grpc client: %s", err)
	}

	return proto.NewTuberClient(conn), conn, nil
}
