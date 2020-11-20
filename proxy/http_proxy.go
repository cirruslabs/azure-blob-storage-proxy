package http_cache

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"log"
	"net"
	"net/http"
)

type StorageProxy struct {
	containerURL  *azblob.ContainerURL
	defaultPrefix string
}

func NewStorageProxy(containerURL *azblob.ContainerURL, defaultPrefix string) *StorageProxy {
	metadataResponse, _ := containerURL.GetProperties(context.Background(), azblob.LeaseAccessConditions{})
	if metadataResponse == nil {
		log.Printf("Creating container %s...", containerURL)
		containerURL.Create(context.Background(), make(map[string]string), azblob.PublicAccessBlob)
	}
	return &StorageProxy{
		containerURL:  containerURL,
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
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	get, err := blockBlobURL.Download(context.Background(), 0, 0, azblob.BlobAccessConditions{}, false)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bufferedReader := bufio.NewReader(get.Body(azblob.RetryReaderOptions{}))
	_, err = bufferedReader.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (proxy StorageProxy) checkBlobExists(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	response, err := blockBlobURL.GetProperties(context.Background(), azblob.BlobAccessConditions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(response.StatusCode())
}

func (proxy StorageProxy) uploadBlob(w http.ResponseWriter, r *http.Request, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))

	_, err := azblob.UploadStreamToBlockBlob(
		context.Background(),
		bufio.NewReader(r.Body),
		blockBlobURL,
		azblob.UploadStreamToBlockBlobOptions{},
	)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusCreated)
}
