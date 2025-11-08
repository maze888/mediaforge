package storage

import (
    "context"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "io"
    "net/url"
    "time"
    "sync"
    "log/slog"
)

var (
    minioClient *minio.Client
    once sync.Once
)

type MinioStorage struct {
    client *minio.Client
    bucket string
} 

func init() {
    // TODO: configuration minio storage connection info
    once.Do(connectMinioClient)
}

// TODO: configuration minio storage connection info
func connectMinioClient() {
    c, err := minio.New("localhost:9000", &minio.Options {
        Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
        Secure: false, // HTTPS 면 true
    })
    if err != nil {
        slog.Error("minio.New failed", "error", err)
        panic(err)
    }
    minioClient = c
}

// GetMinioClient is currently unused but kept for future use.
func GetMinioClient() *minio.Client {
    return minioClient
}

func NewMinioStorage(bucket string) (minioStorage *MinioStorage, err error) {
    exists, err := minioClient.BucketExists(context.Background(), bucket)
    if err != nil {
        return nil, err
    }
    if !exists {
        if err := minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, err
        }
    }

    return &MinioStorage {
        client: minioClient,
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
    return err
}

func (m *MinioStorage) Download(objectName string) (io.ReadCloser, error) {
    return m.client.GetObject(
        context.Background(),
        m.bucket,
        objectName,
        minio.GetObjectOptions {},
        )
}

// GetPresignedUploadURL 보안상 내부 서비스 모듈에서만 사용할 것
func (m *MinioStorage) GetPresignedUploadURL(objectName string, expiry time.Duration) (*url.URL, error) {
    return m.client.PresignedPutObject(
        context.Background(),
        m.bucket,
        objectName,
        expiry,
        )
}

// GetPresignedDownloadURL 보안상 내부 서비스 모듈에서만 사용할 것
func (m *MinioStorage) GetPresignedDownloadURL(objectName string, expiry time.Duration) (*url.URL, error) {
    return m.client.PresignedGetObject(
        context.Background(),
        m.bucket,
        objectName,
        expiry,
        nil,
        )
}
