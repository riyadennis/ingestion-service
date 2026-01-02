package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/riyadennis/identity-server/app/proto/identity"
	"github.com/riyadennis/ingestion-service/business"
	"github.com/riyadennis/ingestion-service/foundation"
	"github.com/sirupsen/logrus"
)

var (
	errInvalidFile  = errors.New("invalid file in the request")
	errFetchingFile = errors.New("error fetching file")
)

type UploadHandler struct {
	Uploader       *business.BucketUpload
	Logger         *logrus.Logger
	identityClient identity.IdentityClient
}

func NewUploader(logger *logrus.Logger, bu *business.BucketUpload, idc identity.IdentityClient) *UploadHandler {
	return &UploadHandler{
		Uploader:       bu,
		Logger:         logger,
		identityClient: idc,
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
	if strings.Contains(contentType, "multipart/form-data") {
		fu, err = business.HandleFormData(w, r)
		if err != nil {
			u.Logger.Errorf("failed to upload form data, got error: %v", err)
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errInvalidFile, foundation.InvalidRequest)
			return
		}

	} else {
		fu, err = business.HandleBinaryData(r)
		if err != nil {
			u.Logger.Errorf("failed to handle binary data, got error: %v", err)
			foundation.ErrorResponse(w, http.StatusBadRequest,
				errInvalidFile, foundation.InvalidRequest)
			return
		}
	}
	userID, err := business.UsrIDFromToken(r.Context(), r, u.identityClient)
	if err != nil {
		u.Logger.Errorf("failed to fetch userID from identity server: %v", err)
		foundation.ErrorResponse(w, http.StatusUnauthorized,
			errFetchingFile, foundation.InvalidRequest)
	}
	fu.UserID = userID

	err = fu.Upload(r.Context(), u.Uploader.Storage, u.Uploader.BucketName)
	if err != nil {
		u.Logger.Errorf("failed to read file: %v", err)
		foundation.ErrorResponse(w, http.StatusInternalServerError,
			errFetchingFile, foundation.InvalidRequest)
	}

}
