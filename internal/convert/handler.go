// Package convert
package convert

import (
    "log/slog"
    
    "mediaforge/storage"

    "github.com/gin-gonic/gin"
)

type FormatRequest struct {
    InputFormat string `form:"inputFormat" json:"inputformat" binding:"required"`
    OutputFormat string `form:"outputFormat" json:"outputformat" binding:"required"`
}

func AddHandler(router *gin.Engine) {
    convertService, err := newConvertService("media")
    if err != nil {
        slog.Error("newConvertService() is failed", "error", err)
        panic(err)
    }

    router.POST("/convert", func(context *gin.Context) {
        convertService.Convert(context)
    })
}
