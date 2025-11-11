// Package convert
package convert

import (
    "fmt"

    "log/slog"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

func AddHandler(router *gin.Engine) {
    convertService, err := newConvertService("media", "amqp://admin:admin123@localhost:5672/", "convert_test")
    if err != nil {
        slog.Error("newConvertService() is failed", "error", err)
        panic(err)
    }

    router.POST("/convert", func(context *gin.Context) {
        var params FormatRequest

        if err := context.ShouldBind(&params); err != nil {
            slog.Error("missing or invalid parameters", "error", err)
            context.JSON(http.StatusBadRequest, gin.H{
                "error": "missing or invalid parameters: " + err.Error(),
            })
            return
        }
        fmt.Printf("params: %v", params)
        // TODO: format parameter validate

        convertRequest, err := convertService.Upload(context, &params)
        fmt.Printf("convertRequest: %v\n", convertRequest)
        if err != nil {
            slog.Error("convertService.Upload() is failed", "error", err)
            context.JSON(http.StatusInternalServerError, gin.H{
                "error": "upload error" + err.Error(),
            })
            return
        }

        if err := convertService.Request(convertRequest); err != nil {
            slog.Error("convertService.Request() is failed", "error", err)
            context.JSON(http.StatusInternalServerError, gin.H{
                "error": "convert request error" + err.Error(),
            })
            return
        }
    })
}
