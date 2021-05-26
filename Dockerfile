FROM bitnami/kubectl:1.15-ol-7

FROM node:16-alpine3.11

ENV TUBER_PREFIX=/tuber
COPY pkg/adminserver/web /app
WORKDIR /app
RUN yarn
RUN yarn build

FROM golang:1.16.4-alpine3.13

COPY --from=0 /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl

RUN mkdir /app
WORKDIR /app

COPY go.mod   ./go.mod
COPY go.sum   ./go.sum
COPY pkg      ./pkg
COPY cmd      ./cmd
COPY main.go  ./main.go
COPY data     ./data
COPY graph    ./graph
COPY .tuber   /.tuber

RUN rm -rf ./pkg/adminserver/web
COPY --from=1 /app/out /static

ENV GO111MODULE on

RUN go build

ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools

CMD ["/app/tuber", "start", "-y"]
