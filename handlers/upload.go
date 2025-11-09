package handlers

import (
	"errors"
	"net/http"

	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/foundation"
	"github.com/sirupsen/logrus"
)

var (
	errFileTooLarge = errors.New("file too large")
	errFetchingFile = errors.New("error fetching file")
)

type UploadHandler struct {
	Uploader *business.BucketUpload
	Logger   *logrus.Logger
}

func NewUploader(logger *logrus.Logger, bu *business.BucketUpload) *UploadHandler {
	return &UploadHandler{
		Uploader: bu,
		Logger:   logger,
	}
}

/*
Upload handles the file upload request
  - File can be uploaded in two ways:
  - 1. multipart/form-data
    Header should contain:
    X-Filename: filename
    Content-Type: image/jpeg, image/png, application/pdf
    body: value file content with key "file"
  - 2. application/octet-stream
    Header should contain:
    X-Filename: filename
    Content-Type: image/jpeg, image/png, application/pdf
    body: file content
*/
func (u *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	u.Logger.Infof("uploading file content type: %s", contentType)
	var (
		fu  *business.FileUpload
		err error
	)
	if contentType == "multipart/form-data" {
		fu, err = business.HandleFormData(w, r, u.Uploader)
		if err != nil {
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errFileTooLarge, foundation.InvalidRequest)
			return
		}

	} else {
		fu, err = business.HandleBinaryData(r, u.Uploader)
		if err != nil {
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errFileTooLarge, foundation.InvalidRequest)
			return
		}
	}
	
	err = fu.Upload(r.Context())
	if err != nil {
		u.Logger.Errorf("failed to read file: %v", err)
		foundation.ErrorResponse(w, http.StatusInternalServerError,
			errFetchingFile, foundation.InvalidRequest)
	}

}
