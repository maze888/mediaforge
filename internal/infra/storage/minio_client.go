// Package storage
package storage

import (
    "fmt"
    "context"
    "time"
    "mime/multipart"
    "mediaforge/internal/util"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
    client *minio.Client
    bucket string // 현재 사용중인 버킷명
}

func NewMinioClient(url, id, secret, bucket string) (minioClient *MinioClient, err error) {
    client, err := minio.New(url, &minio.Options {
        Creds: credentials.NewStaticV4(id, secret, ""),
        Secure: false, // HTTPS
    })
    if err != nil {
        return nil, fmt.Errorf("minio.New() failed: %w", err)
    }

    if err := makeBucket(client, bucket); err != nil {
        return nil, err
    }

    return &MinioClient {
        client: client,
        bucket: bucket,
    }, nil
}

func makeBucket(client *minio.Client, bucket string) error {
    exists, err := client.BucketExists(context.Background(), bucket)
    if err != nil {
        return fmt.Errorf("client.BucketExists() failed: %w", err)
    }
    if !exists {
        if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
            return fmt.Errorf("client.MakeBucket() failed: %w", err)
        }
    }

    return nil
}

func (mc *MinioClient) ChangeBucket(bucket string) error {
    return makeBucket(mc.client, bucket)
}

func (mc *MinioClient) Upload(file *multipart.FileHeader, objectName string) (err error) {
    reader, err := file.Open()
    if err != nil {
        return fmt.Errorf("file.Open() failed: %w", err)
    }

    _, err = mc.client.PutObject(
        context.Background(),
        mc.bucket,
        objectName,
        reader,
        file.Size,
        minio.PutObjectOptions {
            ContentType: util.GetContentType(file),
            UserMetadata: map[string]string {
                "original-filename": file.Filename,
            },
        },
        )
    if err != nil {
        return fmt.Errorf("mc.client.PutObject() failed: %w", err)
    }

    return err
}

func (mc *MinioClient) GetPresignedUploadURL(objectName string, expiry time.Duration) (string, error) {
    url, err := mc.client.PresignedPutObject(
        context.Background(),
        mc.bucket,
        objectName,
        expiry,
        )
    if err != nil {
        return "", fmt.Errorf("mc.client.PresignedPutObject() failed: %w", err)
    }

    return url.String(), err
}

func (mc *MinioClient) GetPresignedDownloadURL(objectName string, expiry time.Duration) (string, error) {
    url, err := mc.client.PresignedGetObject(
        context.Background(),
        mc.bucket,
        objectName,
        expiry,
        nil,
        )
    if err != nil {
        return "", fmt.Errorf("mc.client.PresignedGetObject() failed: %w", err)
    }

    return url.String(), err
}
