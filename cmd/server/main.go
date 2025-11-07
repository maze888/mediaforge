package main

import (
    "github.com/gin-gonic/gin"
    "mediaforge/storage"
    "net/http"
    "log/slog"
)

func main() {
    router := gin.Default()

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"message": "pong"})
    })

    minioStorage, err := storage.NewMinioStorage("localhost:9000", "minioadmin", "minioadmin", "media"); 
    if err != nil {
        slog.Error("Failed to create MinIO storage", "error", err)
        panic(err)
    }

    router.POST("/upload", func(ctx *gin.Context) {
        file, err := ctx.FormFile("file")
        if err != nil {
            return
        }
        
        src, err := file.Open()
        if err != nil {
            return
        }
        defer src.Close()

        if err := minioStorage.Upload(src, file.Filename, file.Size); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusOK, gin.H{"message": "uploaded"})
    })

    router.Run(":5000")
}

