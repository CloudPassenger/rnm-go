FROM golang:1.23 AS build

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0
RUN set -ex \
    && export COMMIT=$(git rev-parse --short HEAD) \
    && export VERSION=$(git rev-parse --abbrev-ref HEAD) \
    && go build -trimpath -ldflags '-s -w -extldflags "-static"' -o /go/bin/rnm-go .


FROM gcr.io/distroless/static-debian12

COPY --from=build /go/bin/rnm-go /usr/bin/

VOLUME /etc/rnm-go

ENTRYPOINT ["rnm-go"]
CMD ["-conf", "/etc/rnm-go/config.json"]
