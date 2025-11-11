package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct {
	uploadInfo minio.UploadInfo
	err        error
}

func (m *MockStorage) FPutObject(_ context.Context, _,
	_, _ string, _ minio.PutObjectOptions) (info minio.UploadInfo, err error) {
	return m.uploadInfo, m.err
}

func TestLoadRESTEndpoints(t *testing.T) {
	scenarios := []struct {
		name               string
		request            *http.Request
		expectedStatusCode int
	}{
		{
			name: "route not found",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", nil)
			}(),
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "liveness",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, LivenessEndPoint, nil)
			}(),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "readiness",
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, ReadinessEndPoint, nil)
			}(),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "upload",
			request: func() *http.Request {
				content := bytes.NewReader([]byte("hello"))
				request := httptest.NewRequest(http.MethodPost, UploadEndpoint, content)
				request.Header.Set("Content-Type", "image/jpeg")
				return request
			}(),
			expectedStatusCode: http.StatusOK,
		},
	}
	logger := logrus.New()
	client := &MockStorage{}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			handler := LoadRESTEndpoints(logger, client, "test")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, scenario.request)
			assert.Equal(t, scenario.expectedStatusCode, w.Code)
		})
	}

}
