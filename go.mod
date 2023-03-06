module github.com/cirruslabs/azure-blob-storage-proxy

go 1.19

replace github.com/cirruslabs/azure-blob-storage-proxy/http_proxy => ./http_proxy

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v0.6.1
	github.com/cirruslabs/azure-blob-storage-proxy/http_proxy v0.0.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.1.4 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.0.1 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
