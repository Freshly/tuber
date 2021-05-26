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

RUN rm -rf ./pkg/adminserver/web/*
COPY --from=1 /app/out ./pkg/adminserver/web/out

RUN ls pkg/adminserver/web/out/_next/static | grep -E '.{21}' | tr -d "\n" > /.folder_name
RUN echo `cat /.folder_name`
RUN sed -i "s/web\/out\/_next\/static\/NEXT_MAGIC_FOLDER_REPLACE/web\/out\/_next\/static\/`cat /.folder_name`/" pkg/adminserver/server.go
RUN rm /.folder_name

ENV GO111MODULE on

RUN go build

CMD ["/app/tuber", "start", "-y"]
