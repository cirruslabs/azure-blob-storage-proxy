package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cirruslabs/azure-blob-storage-proxy/http_proxy"
)

func main() {
	var port int64
	flag.Int64Var(&port, "port", 8080, "Port to serve")
	var defaultPrefix string
	flag.StringVar(&defaultPrefix, "prefix", "", "Optional prefix for all objects. For example, use --prefix=foo/.")
	var containerName string
	flag.StringVar(&containerName, "container", "cirrus-ci-caches", "Container to use for storing caches.")
	var AzureAccountName string
	flag.StringVar(&AzureAccountName, "account-name", "", "Azure Account Name")
	var AzureAccountKey string
	flag.StringVar(&AzureAccountKey, "account-key", "", "Azure Account Key")
	flag.Parse()

	if AzureAccountName == "" {
		log.Fatal("Please specify Azure Account Name")
	}

	if AzureAccountKey == "" {
		log.Fatal("Please specify Azure Account Key")
	}

	credential, err := azblob.NewSharedKeyCredential(AzureAccountName, AzureAccountKey)
	if err != nil {
		log.Fatalf("Failed to create shared credentials: %s", err)
	}
	azureURL := fmt.Sprintf("https://%s.blob.core.windows.net", AzureAccountName)
	client, err := azblob.NewClientWithSharedKeyCredential(azureURL, credential, &azblob.ClientOptions{})
	if err != nil {
		log.Fatalf("Failed to create a storage client: %s", err)
	}
	storageProxy := http_proxy.NewStorageProxy(client, containerName, defaultPrefix)
	err = storageProxy.Serve(port)
	if err != nil {
		log.Fatalf("Failed to start proxy: %s", err)
	}
}
