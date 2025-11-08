package main

import (
    "github.com/gin-gonic/gin"
    "mediaforge/storage"
    "log/slog"
    // "fmt"
    // "time"
    "mediaforge/internal/convert"
)

func main() {
    router := gin.Default()

    // TODO: minio configuration
    minioStorage, err := storage.NewMinioStorage("media"); 
    if err != nil {
        slog.Error("Failed to create MinIO storage", "error", err)
        panic(err)
    }
    
    convert.AddHandler(router, minioStorage)

    router.Run(":5000")
}

