package http_cache

import (
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/go-test/deep"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func CreateLocalProxy(defaultPrefix string) *StorageProxy {
	credential, _ := azblob.NewSharedKeyCredential("devstoreaccount1", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	azureURL, _ := url.Parse("http://localhost:10000/devstoreaccount1")
	serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
	containerURL := serviceURL.NewContainerURL("cirrus-ci-caches-test")

	return NewStorageProxy(&containerURL, defaultPrefix)
}

func Test_All(t *testing.T) {
	expectedBlobContent := "my content"
	storageProxy := CreateLocalProxy("")

	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/some/file", strings.NewReader(expectedBlobContent))
	storageProxy.uploadBlob(response, request, "some/file")

	if response.Code != http.StatusCreated {
		t.Errorf("Wrong status: '%d'", response.Code)
	}

	response = httptest.NewRecorder()
	storageProxy.checkBlobExists(response, "some/file")

	if response.Code != http.StatusOK {
		t.Errorf("Wrong status: '%d'", response.Code)
	}

	response = httptest.NewRecorder()
	storageProxy.downloadBlob(response, "some/file")

	if response.Code != http.StatusOK {
		t.Errorf("Wrong status: '%d'", response.Code)
	}

	downloadedBlobContent := strings.TrimSpace(response.Body.String())
	if diff := deep.Equal(downloadedBlobContent, expectedBlobContent); diff != nil {
		t.Error(len([]byte(downloadedBlobContent)))
		t.Error(len([]byte(expectedBlobContent)))
		t.Error(diff)
	}

	prefixedStorageProxy := CreateLocalProxy("some/")

	response = httptest.NewRecorder()
	prefixedStorageProxy.checkBlobExists(response, "file")

	if response.Code != http.StatusOK {
		t.Errorf("Wrong status: '%d'", response.Code)
	}
}
