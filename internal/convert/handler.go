// Package convert
package convert

import (
    "fmt"
    // "time"

    "log/slog"
    "net/http"
    "encoding/json"

    "github.com/gin-gonic/gin"

    amqp "github.com/rabbitmq/amqp091-go"
)

type ConvertResponse struct {
    JobID string `json:"JobID"`
    UploadFileName string `json:"UploadFileName"`
    ErrorMessage string `json:"ErrorMessage"`
    Status bool `json:"Status"`
}

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
        fmt.Printf("===== params: %+v\n", params)

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

        responses := convertService.Response()

        // CorrID 맞는거 찾을때까지 무한 대기다.. 100% 수신 보장이 될지는 고민좀 해봐야될듯..
        var rejectResponses []*amqp.Delivery
        for response := range responses {
            // fmt.Printf("------ response: %+v\n", response)
            if response.CorrelationId == convertRequest.CorrelationID {
                var convertResponse ConvertResponse

                if err = json.Unmarshal(response.Body, &convertResponse); err != nil {
                    slog.Error("json.Unmarshal(response.Body) is failed", "error: ", err)
                    // TODO: DLX 처리..? Dead Letter
                }
                // fmt.Printf("----- convertResponse: %+v\n", convertResponse)

                if !convertResponse.Status {
                    slog.Error("ffmpegd convert failed", "error: ", convertResponse.ErrorMessage)
                    // context.JSON(http.StatusInternalServerError, gin.H{
                    //     "downloadURL": convertRequest.PresignedUploadURL,
                    // })
                    // TODO: DLX 처리..? Dead Letter
                }

                // 미스매치된 응답들 거부 처리(다른 컨슈머들이 처리할 수 있도록)
                // fmt.Printf("----- mismatched: %v\n", len(rejectResponses))
                for _, rejectResponse := range rejectResponses {
                    if err = rejectResponse.Nack(false, true); err != nil {
                        slog.Error("response.Reuject(true) is failed", "error: ", err)
                    }
                }

                // 요청에 대한 응답만 ACK 처리
                if err = response.Ack(false); err != nil {
                    slog.Error("response.Ack(false) is failed", "error: ", err)
                }
                context.JSON(http.StatusOK, gin.H{
                    "downloadURL": convertRequest.PresignedUploadURL,
                })
                break
            } else {
                rejectResponses = append(rejectResponses, &response)
            }
        }
        // fmt.Printf("-------- POST END\n")
    })
}
