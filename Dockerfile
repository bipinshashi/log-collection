FROM golang:1.21.4 AS build

RUN mkdir /build
WORKDIR /build

# Incrementally COPY and build parts. This ensures cacheable steps.
COPY go.mod .
COPY go.sum .

COPY ./vendor ./vendor
RUN go build -mod=vendor -trimpath -ldflags="-linkmode=internal" ./vendor/...

COPY ./internal ./internal
RUN go build -mod=vendor -trimpath -ldflags="-linkmode=internal" ./internal/...

COPY main.go ./main.go

RUN go build -mod=vendor -trimpath -ldflags="-linkmode=internal" -o /build/log-collection main.go

# Production Image
FROM alpine:3.18.3
RUN apk add --update --no-cache ca-certificates

COPY --from=build /build/log-collection /

ENTRYPOINT ["/log-collection"]
