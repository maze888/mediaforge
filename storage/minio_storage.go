package storage

import (
    "context"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "io"
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
    
func (m *MinioStorage) Upload(reader io.Reader, objectName string, objectSize int64) (err error) {
    _, err = m.client.PutObject(
        context.Background(),
        m.bucket,        // butket name
        objectName,   // object name
        reader,             // stream
        objectSize,       // object size
        minio.PutObjectOptions {
            // ContentType: file.Header.Get("Content-Type"),
        },
    )
    if err != nil {
        return err
    }

    return nil
}
