package main

import (
	"flag"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"github.com/cirruslabs/azure-blob-storage-proxy/proxy"
	"log"
	"net/url"
)

func main() {
	var port int64
	flag.Int64Var(&port, "port", 8080, "Port to serve")
	var defaultPrefix string
	flag.StringVar(&defaultPrefix, "prefix", "", "Optional prefix for all objects. For example, use --prefix=foo/.")
	var containerName string
	flag.StringVar(&defaultPrefix, "container", "cirrus-ci-caches", "Container to use for storing caches.")
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

	credential := azblob.NewSharedKeyCredential(AzureAccountName, AzureAccountKey)
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", AzureAccountName))
	if err != nil {
		log.Fatalf("Failed to create a storage client: %s", err)
	}

	serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
	containerURL := serviceURL.NewContainerURL(containerName)

	storageProxy := http_cache.NewStorageProxy(&containerURL, defaultPrefix)
	err = storageProxy.Serve(port)
	if err != nil {
		log.Fatalf("Failed to start proxy: %s", err)
	}
}
