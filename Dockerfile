FROM golang:latest as builder

WORKDIR /build
ADD . /build

RUN go get -t -v ./... && \
    go build -o azure-blob-storage-proxy ./cmd/

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/cirruslabs/azure-blob-storage-proxy/

WORKDIR /svc
COPY --from=builder /build/azure-blob-storage-proxy /svc/
ENTRYPOINT ["/svc/azure-blob-storage-proxy"]