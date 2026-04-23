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

type Storage interface {
	FPutObject(ctx context.Context, bucketName,
		objectName, filePath string, opts minio.PutObjectOptions) (minio.UploadInfo, error)
}

type BucketUpload struct {
	Storage    Storage
	BucketName string
}

func NewBucketUpload(storage Storage, bucketName string) *BucketUpload {
	return &BucketUpload{
		Storage:    storage,
		BucketName: bucketName,
	}
}

type FileUploader interface {
	Upload(ctx context.Context, storage Storage, bucketName string) error
}

type FileUpload struct {
	RealName    string
	FileName    string
	File        io.Reader
	Size        int64
	ContentType string
	UserID      string
}

/*
Upload uploads a file to storage client set on start up
  - Validates file type (JPEG, PNG, PDF, octet-stream)
  - Reads file data (max 100MB)
  - Generates safe filename with random hex string
  - Writes temporary file
  - Uploads to MinIO bucket
  - Cleans up temp file
*/
func (f *FileUpload) Upload(ctx context.Context, storage Storage, bucketName string) error {
	// validate file type
	if !AllowedTypes[f.ContentType] {
		return errUnsupportedFileType
	}
	// save a temporary copy of the file
	data := make([]byte, f.Size)
	if _, err := io.ReadFull(f.File, data); err != nil {
		return err
	}
	generatedFileName := generateSafeFilename(f.FileName, f.ContentType)
	if err := os.WriteFile(generatedFileName, data, 0644); err != nil {
		return err
	}

	var removeErr error
	defer func() {
		err := os.Remove(generatedFileName)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			removeErr = err
		}
	}()
	if removeErr != nil {
		return removeErr
	}
	// upload the file to the bucket
	_, err := storage.FPutObject(ctx,
		bucketName, generatedFileName, generatedFileName,
		minio.PutObjectOptions{
			ContentType: f.ContentType,
			UserMetadata: map[string]string{
				"fileName": f.RealName,
				"userID":   f.UserID,
			},
		})
	if err != nil {
		return err
	}

	return nil
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
