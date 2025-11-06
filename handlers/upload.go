package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/riyadennis/ingestion-service/foundation"
	"github.com/sirupsen/logrus"
)

var (
	errFileTooLarge        = errors.New("file too large")
	errFetchingFile        = errors.New("error fetching file")
	errUnsupportedFileType = errors.New("unsupported file type")
	errFailedToUpload      = errors.New("failed to upload file")
)

type Uploader struct {
	Storage      *minio.Client
	BucketName   string
	Logger       *logrus.Logger
	MaxFileSize  int64
	AllowedTypes map[string]bool
}

func NewUploader(storage *minio.Client, logger *logrus.Logger, bucketName string) *Uploader {
	return &Uploader{
		Storage:     storage,
		BucketName:  bucketName,
		Logger:      logger,
		MaxFileSize: 1024 * 1024 * 100,
		AllowedTypes: map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"application/pdf": true,
		},
	}
}

func (u *Uploader) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(u.MaxFileSize)
	if err != nil {
		foundation.ErrorResponse(w, http.StatusBadRequest,
			errFileTooLarge, foundation.InvalidRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, u.MaxFileSize)
	file, header, err := r.FormFile("file")
	if err != nil {
		foundation.ErrorResponse(w, http.StatusBadRequest,
			errFileTooLarge, foundation.InvalidRequest)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	err = u.validateFileType(header)
	if err != nil {
		u.Logger.Errorf("validation of file failed: %v", err)
		foundation.ErrorResponse(w, http.StatusBadRequest,
			errUnsupportedFileType, foundation.ValidationFailed)
		return
	}

	data := make([]byte, header.Size)
	_, err = io.ReadFull(file, data)
	if err != nil {
		u.Logger.Errorf("failed to read file: %v", err)
		foundation.ErrorResponse(w, http.StatusInternalServerError,
			errFetchingFile, foundation.InvalidRequest)
	}
	fileName := generateSafeFilename(header.Filename)
	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		u.Logger.Errorf("failed to write temporary file: %v", err)
		foundation.ErrorResponse(w, http.StatusInternalServerError,
			errFailedToUpload, foundation.InvalidRequest)
		return
	}
	_, err = u.Storage.FPutObject(r.Context(),
		u.BucketName, fileName, fileName, minio.PutObjectOptions{
			ContentType: r.Header.Get("Content-Type"),
		})
	if err != nil {
		u.Logger.Errorf("failed to upload file: %v", err)
		foundation.ErrorResponse(w, http.StatusInternalServerError,
			errFailedToUpload, foundation.InvalidRequest)
		return
	}
	_ = os.Remove(fileName)
}

func (u *Uploader) validateFileType(header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if !u.AllowedTypes[contentType] {
		return errUnsupportedFileType
	}
	return nil
}

func generateSafeFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return originalName
	}
	return hex.EncodeToString(randomBytes) + ext
}
