package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/foundation"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	scenarios := []struct {
		name             string
		request          *http.Request
		storage          business.Storage
		expectedStatus   int
		expectedResponse string
	}{
		{
			name: "invalid file in form",
			request: func() *http.Request {
				content := bytes.NewReader([]byte("hello"))
				request := httptest.NewRequest(http.MethodPost, UploadEndpoint, content)
				request.Header.Set("Content-Type", "multipart/form-data")
				return request
			}(),
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: errInvalidFile.Error(),
		},
		{
			name: "unsupported content type",
			request: func() *http.Request {
				content := bytes.NewReader([]byte("hello"))
				request := httptest.NewRequest(http.MethodPost, UploadEndpoint, content)
				request.Header.Set("Content-Type", "INVALID")
				return request
			}(),
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: errInvalidFile.Error(),
		},
		{
			name: "upload failed",
			request: func() *http.Request {
				content := bytes.NewReader([]byte("hello"))
				request := httptest.NewRequest(http.MethodPost, UploadEndpoint, content)
				request.Header.Set("Content-Type", "image/jpeg")
				return request
			}(),
			storage: &MockStorage{
				err: errors.New("failed to upload"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success",
			request:        requestWithFile(t),
			storage:        &MockStorage{},
			expectedStatus: http.StatusOK,
		},
	}
	logger := logrus.New()
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			up := NewUploader(logger, business.NewBucketUpload(scenario.storage, "test"))
			w := httptest.NewRecorder()
			up.Upload(w, scenario.request)
			data, err := io.ReadAll(w.Result().Body)
			assert.NoError(t, err)
			if scenario.expectedResponse != "" {
				res := &foundation.Response{}
				err = json.Unmarshal(data, res)
				assert.NoError(t, err)
				assert.Equal(t, scenario.expectedStatus, w.Code)
				assert.Equal(t, scenario.expectedResponse, res.Message)
			}
		})

	}
}

func requestWithFile(t *testing.T) *http.Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	file, err := os.Open("../testdata/test.txt")
	assert.NoError(t, err)
	defer func() {
		err = file.Close()
		assert.NoError(t, err)
	}()
	fileW, err := writer.CreateFormFile("file", "../testdata/test.txt")
	assert.NoError(t, err)
	_, err = io.Copy(fileW, file)
	assert.NoError(t, err)
	defer func() {
		err = writer.Close()
		assert.NoError(t, err)
	}()

	req := httptest.NewRequest(http.MethodPost,
		"/request", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}
