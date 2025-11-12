// Package convert
package convert

import (
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
        // TODO: 고루틴 처리 (context 는 복사해야함)
        var params FormatRequest

        if err := context.ShouldBind(&params); err != nil {
            slog.Error("missing or invalid parameters", "error", err)
            context.JSON(http.StatusBadRequest, gin.H{
                "error": "missing or invalid parameters: " + err.Error(),
            })
            return
        }
        // TODO: format parameter validate

        convertRequest, err := convertService.Upload(context, &params)
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

        // TODO: 완료 메시지 수신 consume
    })
}
