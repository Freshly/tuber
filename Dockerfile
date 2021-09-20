FROM bitnami/kubectl:1.19
FROM golang:1.16.5-alpine3.13

COPY --from=0 /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl

RUN mkdir /app
WORKDIR /app

COPY go.mod   ./go.mod
COPY go.sum   ./go.sum
RUN go mod download

COPY pkg      ./pkg
COPY cmd      ./cmd
COPY main.go  ./main.go
COPY data     ./data
COPY graph    ./graph
COPY .tuber   /.tuber

RUN go build

CMD ["/app/tuber", "start", "-y"]
