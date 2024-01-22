FROM golang:1.18 as build
RUN set -ex && \
    go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn && \
    go env -w GOSUMDB=off && \
    go env

ENV GO111MODULE=on \
    GOPROXY="https://goproxy.cn,direct" \
    GOSUMDB=off \
    GOARCH=amd64 \
    CGO_ENABLED=0

WORKDIR /go/src/github.com/bnulwh/go-selfupdate/
COPY ./ /go/src/github.com/bnulwh/go-selfupdate/
RUN  pwd && \
     ls -lh && \
     rm -fr server \
     rm -fr go-selfupdate \
     go mod vendor -v && \
     go build  -v -ldflags -s -a -installsuffix cgo -o  server ./cmd/server/

FROM 5ibnu/centos:7.4
#COPY ./resources /resources
#ADD  ./application.properties /
COPY --from=build  /go/src/github.com/bnulwh/go-selfupdate/server /
EXPOSE 8080
ENTRYPOINT ["/server"]

