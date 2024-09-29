FROM golang:1.21.4 AS build

RUN mkdir /build
WORKDIR /build

WORKDIR /go/src/github.com/bipinshashi/log-collection
COPY go.mod .
COPY go.sum .

COPY ./vendor ./vendor
RUN go build -mod=vendor -trimpath -ldflags="-linkmode=internal" ./vendor/...

COPY ./internal ./internal
RUN go build -mod=vendor -trimpath -ldflags="-linkmode=internal" ./internal/...

COPY main.go ./main.go

RUN set -x && \
    CGO_ENABLED=0 go build -mod=vendor \
        -o /build/log-collection

# Production Image
FROM alpine:3.18.3
RUN apk add --update --no-cache ca-certificates

WORKDIR /root/
COPY --from=build /build/log-collection .

ENTRYPOINT [ "./log-collection" ]
