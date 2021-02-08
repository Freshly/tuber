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
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := grpc.Dial(url+":443", grpc.WithTransportCredentials(credentials.NewTLS(config)))

	if err != nil {
		return nil, nil, fmt.Errorf("grpc client: %s", err)
	}

	return proto.NewTuberClient(conn), conn, nil
}
