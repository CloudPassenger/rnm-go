FROM golang:1.20-alpine AS builder
COPY . /go/src/github.com/CloudPassenger/rmm-go
WORKDIR /go/src/github.com/CloudPassenger/rmm-go

ENV CGO_ENABLED=0
RUN set -ex \
    && apk add git build-base \
    && export COMMIT=$(git rev-parse --short HEAD) \
    && export VERSION=$(git rev-parse --abbrev-ref HEAD) \
    && go build -ldflags '-s -w -extldflags "-static"' -o rnm-go .

FROM alpine AS dist
COPY --from=builder /go/src/github.com/CloudPassenger/rmm-go/rnm-go /usr/bin/
VOLUME /etc/rnm-go
ENTRYPOINT ["rnm-go", "-conf", "/etc/rnm-go/config.json"]
