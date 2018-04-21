package http_cache

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"io"
	"log"
	"net"
	"net/http"
)

type StorageProxy struct {
	containerURL  *azblob.ContainerURL
	defaultPrefix string
}

func NewStorageProxy(containerURL *azblob.ContainerURL, defaultPrefix string) *StorageProxy {
	metadataResponse, _ := containerURL.GetPropertiesAndMetadata(context.Background(), azblob.LeaseAccessConditions{})
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

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))

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
	downloadStream := azblob.NewDownloadStream(
		context.Background(),
		func(ctx context.Context, blobRange azblob.BlobRange, ac azblob.BlobAccessConditions, rangeGetContentMD5 bool) (*azblob.GetResponse, error) {
			return blockBlobURL.GetBlob(ctx, blobRange, ac, rangeGetContentMD5)
		},
		azblob.DownloadStreamOptions{},
	)
	defer downloadStream.Close()
	bufferedReader := bufio.NewReader(downloadStream)
	_, err := bufferedReader.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (proxy StorageProxy) checkBlobExists(w http.ResponseWriter, name string) {
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	response, err := blockBlobURL.GetPropertiesAndMetadata(context.Background(), azblob.BlobAccessConditions{})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(response.StatusCode())
}

func (proxy StorageProxy) uploadBlob(w http.ResponseWriter, r *http.Request, name string) {
	blockIDIntToBase64 := func(blockID int) string {
		binaryBlockID := (&[4]byte{})[:] // All block IDs are 4 bytes long
		binary.LittleEndian.PutUint32(binaryBlockID, uint32(blockID))
		return base64.StdEncoding.EncodeToString(binaryBlockID)
	}
	blockBlobURL := proxy.containerURL.NewBlockBlobURL(proxy.objectName(name))
	readBufferSize := int(1024 * 1024)
	readBuffer := make([]byte, readBufferSize)
	bufferedBodyReader := bufio.NewReaderSize(r.Body, readBufferSize)
	uploadedParts := 0
	blockIds := make([]string, 0)
	for {
		n, err := bufferedBodyReader.Read(readBuffer)

		if n > 0 {
			blockId := blockIDIntToBase64(uploadedParts)
			_, err = blockBlobURL.PutBlock(
				context.Background(),
				blockId,
				bytes.NewReader(readBuffer[:n]),
				azblob.LeaseAccessConditions{},
			)
			blockIds = append(blockIds, blockId)
			uploadedParts += 1
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			errorMsg := fmt.Sprintf("Failed read cache body! %s", err)
			log.Print(errorMsg)
			w.Write([]byte(errorMsg))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	_, err := blockBlobURL.PutBlockList(
		context.Background(),
		blockIds,
		azblob.Metadata{},
		azblob.BlobHTTPHeaders{},
		azblob.BlobAccessConditions{},
	)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusCreated)
}
