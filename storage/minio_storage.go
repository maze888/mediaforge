package storage

import (
    "context"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "github.com/gin-gonic/gin"
)

type MinioStorage struct {
    client *minio.Client
    bucket string
} 

func NewMinioStorage(endpoint, accessKey, secretKey, bucket string) (*MinioStorage, error) {
    client, err := minio.New(endpoint, &minio.Options {
        Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: false, // HTTPS 면 true
    })
    if err != nil {
        return nil, err
    }

    // 버킷 없으면 생성
    exists, err := client.BucketExists(context.Background(), bucket)
    if err != nil {
        return nil, err
    }
    if !exists {
        if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, err
        }
    }

    return &MinioStorage {
        client: client,
        bucket: bucket,
    }, nil
}
    
func (m *MinioStorage) Upload(ctx *gin.Context) (err error) {
    file, err := ctx.FormFile("file")
    if err != nil {
        return err
    }
    
    src, err := file.Open()
    if err != nil {
        return err
    }
    defer src.Close()

    _, err = m.client.PutObject(
        context.Background(),
        m.bucket,        // butket name
        file.Filename,   // object name
        src,             // stream
        file.Size,       // object size
        minio.PutObjectOptions {
            ContentType: file.Header.Get("Content-Type"),
        },
    )
    if err != nil {
        return err
    }

    return nil
}
