// Package storage provides interfaces and implementations
// for object storage systems like MinIO and S3.
package storage

import (
    "io"
    "time"
    "net/url"
)

type ObjectStorage interface {
    Upload(reader io.Reader, objectName string, objectSize int64) error
    Download(objectName string) (io.ReadCloser, error)
    GetPresignedUploadURL(objectName string, expiry time.Duration) (*url.URL, error)
    GetPresignedDownloadURL(objectName string, expiry time.Duration) (*url.URL, error)
}

