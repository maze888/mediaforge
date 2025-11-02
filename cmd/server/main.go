package main

import (
    "github.com/gin-gonic/gin"
    "mediaforge/storage"
    "net/http"
    "log"
)

func main() {
    router := gin.Default()

    router.GET("/ping", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"message": "pong"})
    })

    minioStorage, err := storage.NewMinioStorage("localhost:9000", "minioadmin", "minioadmin", "media"); 
    if err != nil {
        log.Fatal(err)
    }

    router.POST("/upload", func(ctx *gin.Context) {
        if err := minioStorage.Upload(ctx); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        ctx.JSON(http.StatusOK, gin.H{"message": "uploaded"})
    })

    router.Run(":5000")
}

