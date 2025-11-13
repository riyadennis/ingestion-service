package business

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleFormData(t *testing.T) {
	scenarios := []struct {
		name           string
		request        *http.Request
		expectedError  error
		expectedResult *FileUpload
	}{
		{
			name: "invalid content type",
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				req.Header.Set("Content-Type", "INVALID")
				return req
			}(),
			expectedError: errInvalidRequest,
		},
		{
			name:    "success",
			request: requestWithFile(t),
			expectedResult: &FileUpload{
				FileName:    "test.txt",
				ContentType: "application/octet-stream",
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			fu, err := HandleFormData(w, scenario.request)
			assert.Equal(t, scenario.expectedError, err)
			if fu != nil {
				assert.Equal(t, scenario.expectedResult.FileName, fu.FileName)
				assert.Equal(t, scenario.expectedResult.ContentType, fu.ContentType)
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
