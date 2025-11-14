package business

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

var errInvalidRequest = errors.New("invalid request")

func HandleFormData(w http.ResponseWriter, r *http.Request) (*FileUpload, error) {
	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		return nil, errInvalidRequest
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxFileSize)
	// key in the request body should be file
	formFile, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = formFile.Close()
	}()

	return &FileUpload{
		FileName:    sanitizeFilename(header.Filename),
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		File:        formFile,
	}, nil
}

func HandleBinaryData(r *http.Request) (*FileUpload, error) {
	if !AllowedTypes[r.Header.Get("Content-Type")] {
		return nil, errUnsupportedFileType
	}

	var requestBody bytes.Buffer
	_, err := io.Copy(&requestBody, r.Body)
	if err != nil {
		return nil, err
	}

	return &FileUpload{
		// get the file name from the header which should be X-Filename
		FileName:    sanitizeFilename(r.Header.Get("X-Filename")),
		Size:        int64(requestBody.Len()),
		ContentType: r.Header.Get("Content-Type"),
		File:        bytes.NewReader(requestBody.Bytes()),
	}, nil
}

func sanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "\x00", "")
	filename = strings.ReplaceAll(filename, "..", "")
	return filepath.Base(filename)
}
