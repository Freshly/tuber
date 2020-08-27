build:
  go build

protoc:
  cd pkg/proto && protoc --go_out=plugins=grpc:. server.proto
