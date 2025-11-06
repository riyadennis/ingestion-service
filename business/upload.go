package business

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
)

var (
	errUnsupportedFileType = errors.New("unsupported file type")
	AllowedTypes           = map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
)

type BucketUpload struct {
	Storage      *minio.Client
	BucketName   string
	MaxFileSize  int64
	AllowedTypes map[string]bool
}

func NewBucketUpload(storage *minio.Client, bucketName string) *BucketUpload {
	return &BucketUpload{
		Storage:     storage,
		BucketName:  bucketName,
		MaxFileSize: 1024 * 1024 * 100,
	}
}

type FileUploadHandler interface {
	Upload(ctx context.Context, file io.Reader) error
}

type FileUpload struct {
	Uploader    *BucketUpload
	FileName    string
	Size        int64
	ContentType string
}

func NewFileUpload(uploader *BucketUpload, fileName string, size int64, contentType string) *FileUpload {
	return &FileUpload{
		Uploader:    uploader,
		FileName:    fileName,
		Size:        size,
		ContentType: contentType,
	}
}

func (f *FileUpload) Upload(ctx context.Context, file io.Reader) error {
	// save a temporary copy of the file
	data := make([]byte, f.Size)
	_, err := io.ReadFull(file, data)
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

func ValidateFileType(header *multipart.FileHeader) error {
	contentType := header.Header.Get("Content-Type")
	if !AllowedTypes[contentType] {
		return errUnsupportedFileType
	}
	return nil
}
