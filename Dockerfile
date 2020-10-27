FROM bitnami/kubectl:1.15-ol-7

FROM golang:1.13.5-alpine3.10

COPY --from=0 /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl
ADD https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 /usr/local/bin/kubectl-argo-rollouts
RUN chmod +x /usr/local/bin/kubectl-argo-rollouts
ENV PATH="/kubectl-argo-rollouts-darwin-amd64:${PATH}"

RUN mkdir /app
WORKDIR /app

COPY go.mod   ./go.mod
COPY go.sum   ./go.sum
COPY pkg      ./pkg
COPY cmd      ./cmd
COPY main.go  ./main.go
COPY data     ./data
COPY .tuber   /.tuber

ENV GO111MODULE on

RUN go build

CMD ["/app/tuber", "start", "-y"]
