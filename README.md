[![Build Status](https://api.cirrus-ci.com/github/cirruslabs/azure-blob-storage-proxy.svg)](https://cirrus-ci.com/github/cirruslabs/azure-blob-storage-proxy) [![](https://images.microbadger.com/badges/image/cirrusci/azure-blob-storage-proxy.svg)](https://microbadger.com/images/cirrusci/azure-blob-storage-proxy)

HTTP proxy with REST API to interact with Azure's Blob Storage.

Simply allows to use `HEAD`, `GET` or `PUT` requests to check blob's availability, as well as downloading or uploading
blobs to a specified Azure container by blob's name.

For example, `GET` for `<proxy_url>/some/file` will return `some/file` blob if it exists.

# Testing

Tests expect to have Azure API available on `localhost:10000`. It is recommended to run [`azurite`](https://github.com/azure/azurite) like this:

```bash
docker run -d -t -p 10000:10000 -p 10001:10001 arafato/azurite
```

After azurite is up and running on `10000` port simply run `go test ./...` to test all the things.