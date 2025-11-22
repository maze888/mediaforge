package main

import (
    "net/http"
    "github.com/gin-gonic/gin"

    "mediaforge2/internal/convert"
    "mediaforge2/internal/infra/queue"
    "mediaforge2/internal/infra/storage"
)

func initHandler(cs *convert.ConvertService) {
    if err := cs.AddHandler(http.MethodPost, "/convert", convert.ConvertFile); err != nil {
        panic(err)
    }
}

func main() {
    router := gin.Default()

    // 내부 모듈 초기화
    rabbitmqClient, err := queue.NewRabbitmqClient("amqp://admin:admin123@localhost:5672", "convert_request", "convert_response")
    if err != nil {
        panic(err)
    }
    minioClient, err := storage.NewMinioClient("localhost:9000", "minioadmin", "minioadmin", "media")
    if err != nil {
        panic(err)
    }
    convertService := convert.NewConvertService(router, rabbitmqClient, minioClient)

    // RestAPI 핸들러 초기화
    initHandler(convertService)

    router.Run(":5000")
}
