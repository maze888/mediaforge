// Package convert
package convert

import (
    "fmt"

    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
)

func AddHandler(router *gin.Engine) {
    convertService, err := newConvertService("media", "amqp://admin:admin123@localhost:5672/", "convert_request", "convert_response")
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
        // TODO: format parameter validate
        // fmt.Printf("%v", params)

        convertRequest, err := convertService.Upload(context, &params)
        if err != nil {
            slog.Error("convertService.Upload() is failed", "error: ", err)
            return
        }
        convertRequest.JobID = params.JobID

        // TODO: 에러시 HTTP JSON 응답 처리
        if err = convertService.Request(convertRequest); err != nil {
            slog.Error("convertService.Request() is failed", "error: ", err)
            return
        }

        responses, err := convertService.Response()
        if err != nil {
            slog.Error("convertService.Response() is failed", "error: ", err)
            return
        }

        // CorrID 맞는거 찾을때까지 무한 대기다.. 100% 수신 보장이 될지는 고민좀 해봐야될듯..
        for response := range responses {
            fmt.Printf("------ response: %+v\n", response)
            if response.CorrelationId == convertRequest.CorrelationID {
                // 요청에 대한 응답만 ACK 처리
                if err = response.Ack(false); err != nil {
                    slog.Error("response.Ack(false) is failed", "error: ", err)
                }
                context.JSON(http.StatusOK, gin.H{
                    "msg": "success",
                })
                return
            } else {
                // 요청에 대한 응답이 아닌 경우는 requeuing
                if err = response.Reject(true); err != nil {
                    slog.Error("response.Reuject(true) is failed", "error: ", err)
                }
            }
        }
    })
}
