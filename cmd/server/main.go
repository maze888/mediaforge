package main

import (
    "mediaforge/internal/convert"

    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()

    convert.AddHandler(router)

    router.Run(":5000")
}

