package handlers

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/foundation"
	"github.com/sirupsen/logrus"
)

var (
	errFileTooLarge = errors.New("file too large")
	errFetchingFile = errors.New("error fetching file")

	errFailedToUpload = errors.New("failed to upload file")
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

func (u *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	u.Logger.Infof("uploading file content type: %s", contentType)
	var (
		fu   *business.FileUpload
		file io.Reader
	)
	if contentType == "multipart/form-data" {
		err := r.ParseMultipartForm(u.Uploader.MaxFileSize)
		if err != nil {
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errFileTooLarge, foundation.InvalidRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, u.Uploader.MaxFileSize)
		formFile, header, err := r.FormFile("file")
		if err != nil {
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errFileTooLarge, foundation.InvalidRequest)
			return
		}
		defer func() {
			_ = formFile.Close()
		}()
		file = formFile
		// validate file type
		err = business.ValidateFileType(header)
		if err != nil {
			u.Logger.Errorf("validation of file failed: %v", err)
			foundation.ErrorResponse(w, http.StatusBadRequest,
				err, foundation.ValidationFailed)
			return
		}
		fu = business.NewFileUpload(
			u.Uploader,
			header.Filename,
			header.Size,
			header.Header.Get("Content-Type"))

	}
	for fType, _ := range business.AllowedTypes {
		if contentType == fType ||
			contentType == "application/octet-stream" {
			filename := r.Header.Get("X-Filename")
			var requestBody bytes.Buffer
			_, err := io.Copy(&requestBody, r.Body)
			if err != nil {
				u.Logger.Errorf("failed calcaulate size of the file: %v", err)
				foundation.ErrorResponse(w, http.StatusInternalServerError,
					errFetchingFile, foundation.InvalidRequest)
				return
			}
			file = bytes.NewReader(requestBody.Bytes())
			fu = business.NewFileUpload(
				u.Uploader,
				filename,
				int64(requestBody.Len()),
				contentType)
		}
	}
	if fu != nil && file != nil {
		err := fu.Upload(r.Context(), file)
		if err != nil {
			u.Logger.Errorf("failed to read file: %v", err)
			foundation.ErrorResponse(w, http.StatusInternalServerError,
				errFetchingFile, foundation.InvalidRequest)
		}
	}

}
