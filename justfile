run *args:
    go run main.go {{args}}
build:
  go build
protoc:
  protoc --go_opt=paths=source_relative --go_out=plugins=grpc:. pkg/proto/tuber_service.proto
