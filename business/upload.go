package business

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
)

const MaxFileSize = 1024 * 1024 * 100

var (
	errUnsupportedFileType = errors.New("unsupported file type")
	AllowedTypes           = map[string]bool{
		"image/jpeg":               true,
		"image/png":                true,
		"application/pdf":          true,
		"application/octet-stream": true,
	}
)

type BucketUpload struct {
	Storage    *minio.Client
	BucketName string
}

func NewBucketUpload(storage *minio.Client, bucketName string) *BucketUpload {
	return &BucketUpload{
		Storage:    storage,
		BucketName: bucketName,
	}
}

type FileUploader interface {
	Upload(ctx context.Context) error
}

type FileUpload struct {
	Uploader    *BucketUpload
	FileName    string
	Size        int64
	ContentType string
	File        io.Reader
}

func (f *FileUpload) Upload(ctx context.Context) error {
	// validate file type
	if !AllowedTypes[f.ContentType] {
		return errUnsupportedFileType
	}
	// save a temporary copy of the file
	data := make([]byte, f.Size)
	_, err := io.ReadFull(f.File, data)
	if err != nil {
		return err
	}
	generatedFileName := generateSafeFilename(f.FileName, f.ContentType)
	err = os.WriteFile(generatedFileName, data, 0644)
	if err != nil {
		return err
	}
	// upload the file to the bucket
	_, err = f.Uploader.Storage.FPutObject(ctx,
		f.Uploader.BucketName, generatedFileName, generatedFileName,
		minio.PutObjectOptions{
			ContentType: f.ContentType,
		})
	if err != nil {
		return err
	}
	// delete the temporary file
	return os.Remove(generatedFileName)
}

func generateSafeFilename(originalName, contentType string) string {
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = "jpeg"
	case "image/png":
		ext = "png"
	case "application/pdf":
		ext = "pdf"
	default:
		ext = "doc"
	}
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return originalName
	}
	return hex.EncodeToString(randomBytes) + "." + ext
}
