// Package storage provides interfaces and implementations
// for object storage systems like MinIO and S3.
package storage
    
import "github.com/gin-gonic/gin"

type ObjectStorage interface {
    Upload(ctx *gin.Context) error
    // Download() error
    // Delete() error
}

