token:
  go run main.go access-token

start:
  go run main.go start

install:
  kubectl apply -f bootstrap.yaml

build:
  go build && mv tuber ~/.bin