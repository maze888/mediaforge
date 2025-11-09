package main

import (
    "github.com/gin-gonic/gin"
    // "log/slog"
    // "fmt"
    // "time"
    "mediaforge/internal/convert"
)

func main() {
    router := gin.Default()

    convert.AddHandler(router)

    router.Run(":5000")
}

