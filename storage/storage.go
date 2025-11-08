// Package storage provides interfaces and implementations
// for object storage systems like MinIO and S3.
package storage

import "io"
    
type ObjectStorage interface {
    Upload(reader io.Reader, objectName string, objectSize int64) error
    Download() (io.ReadCloser, error)
    // Delete() error
}

