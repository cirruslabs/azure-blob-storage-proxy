package http_proxy

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type StorageProxy struct {
	client        *azblob.Client
	containerName string
	defaultPrefix string
}

func NewStorageProxy(client *azblob.Client, containerName string, defaultPrefix string) *StorageProxy {
	metadataResponse, _ := client.ServiceClient().NewContainerClient(containerName).GetProperties(context.Background(), &container.GetPropertiesOptions{})
	if metadataResponse.Metadata == nil {
		log.Printf("Creating container %s...", containerName)
		client.CreateContainer(context.Background(), containerName, &container.CreateOptions{})
	}

	return &StorageProxy{
		client:        client,
		containerName: containerName,
		defaultPrefix: defaultPrefix,
	}
}

func (proxy StorageProxy) objectName(name string) string {
	return proxy.defaultPrefix + name
}

func (proxy StorageProxy) Serve(port int64) error {
	http.HandleFunc("/", proxy.handler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err == nil {
		address := listener.Addr().String()
		listener.Close()
		log.Printf("Starting http cache server %s\n", address)
		return http.ListenAndServe(address, nil)
	}
	return err
}

func (proxy StorageProxy) handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if key[0] == '/' {
		key = key[1:]
	}
	if r.Method == "GET" {
		proxy.downloadBlob(w, key)
	} else if r.Method == "HEAD" {
		proxy.checkBlobExists(w, key)
	} else if r.Method == "POST" {
		proxy.uploadBlob(w, r, key)
	} else if r.Method == "PUT" {
		proxy.uploadBlob(w, r, key)
	}
}

func (proxy StorageProxy) downloadBlob(w http.ResponseWriter, name string) {
	streamResponse, err := proxy.client.DownloadStream(context.Background(), proxy.containerName, proxy.objectName(name), &azblob.DownloadStreamOptions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	bufferedReader := bufio.NewReader(streamResponse.Body)
	_, err = bufferedReader.WriteTo(w)
	if err != nil {
		log.Printf("Failed to serve blob %q: %v", name, err)
	}
}

func (proxy StorageProxy) checkBlobExists(w http.ResponseWriter, name string) {
	blobClient := proxy.client.ServiceClient().NewContainerClient(proxy.containerName).NewBlobClient(proxy.objectName(name))
	_, err := blobClient.GetProperties(context.Background(), &blob.GetPropertiesOptions{})

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (proxy StorageProxy) uploadBlob(w http.ResponseWriter, r *http.Request, name string) {
	_, err := proxy.client.UploadStream(context.Background(), proxy.containerName, proxy.objectName(name), bufio.NewReader(r.Body), &azblob.UploadStreamOptions{})
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusCreated)
}
