FROM golang:1.10-alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git

WORKDIR /go/src/github.com/cirruslabs/azure-blob-storage-proxy
ADD . /go/src/github.com/cirruslabs/azure-blob-storage-proxy

RUN go get -t -v ./... && \
    go build -o azure-blob-storage-proxy ./cmd/

FROM alpine
WORKDIR /svc
COPY --from=builder /go/src/github.com/cirruslabs/azure-blob-storage-proxy/azure-blob-storage-proxy /svc/
ENTRYPOINT ["/svc/azure-blob-storage-proxy"]