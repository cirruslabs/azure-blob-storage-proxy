module github.com/cirruslabs/azure-blob-storage-proxy

go 1.19

replace github.com/cirruslabs/azure-blob-storage-proxy/http_proxy => ./http_proxy

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.2
	github.com/cirruslabs/azure-blob-storage-proxy/http_proxy v0.0.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.12.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.9.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)
