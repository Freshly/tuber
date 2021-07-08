run *args:
  go run main.go {{args}}

build:
  go build

protoc:
  protoc --go_opt=paths=source_relative --go_out=plugins=grpc:. pkg/proto/tuber_service.proto

gen:
  go generate ./...
  cd pkg/adminserver/web && yarn generate

web *args:
  cd pkg/adminserver/web && yarn {{args}}

local-image:
  docker build . -t tuber
  docker run \
    --rm \
    -it \
    --name tuber \
    --env-file .env \
    --expose $TUBER_ADMINSERVER_PORT \
    -p $TUBER_ADMINSERVER_PORT:$TUBER_ADMINSERVER_PORT \
    -v $HOME/.kube:/root/.kube \
    -v $HOME/.config/gcloud:/root/.config/gcloud \
    -v /usr/lib/google-cloud-sdk:/usr/lib/google-cloud-sdk \
    tuber \
    /app/tuber adminserver -y

remove-tag *version:
  git tag -d {{version}}
  git push origin :{{version}}
