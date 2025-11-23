// Package convert
package convert

import (
    "fmt"
    "time"
    "log/slog"
    "net/http"
    "mime/multipart"
    "encoding/json"

    "mediaforge/internal/util"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// ClientRequest from web client
type ClientRequest struct {
    JobID []string `form:"JobID" binding:"required"`
    InputFormat []string `form:"InputFormat" binding:"required"`
    OutputFormat []string `form:"OutputFormat" binding:"required"`
    ConversionFiles []*multipart.FileHeader `form:"files" binding:"required"`
}

// JobResponse To WebClient
type JobResponse struct {
    JobID string `json:"job_id" binding:"required"`
    DownloadURL string `json:"download_url" binding:"required"`
}

// ConvertRequest to ffmpegd (one by one)
type ConvertRequest struct {
    JobID string
    FileName string
    InputFormat, OutputFormat string
    PresignedUploadURL string
    PresignedDownloadURL string
}

// ConvertResponse from ffmpegd (one by one)
type ConvertResponse struct {
    JobID string `json:"job_id" binding:"required"`
    ErrorMessage string `json:"error_message" binding:"required"`
    Status bool `json:"status" binding:"required"`
}

func getConvertedFileName(clientRequest *ClientRequest, jobID string) (name string) {
    for i, id := range clientRequest.JobID {
        if id == jobID {
            name = fmt.Sprintf("%s.%s", util.GetFileName(clientRequest.ConversionFiles[i].Filename), clientRequest.OutputFormat[i])
        }
    }
    return name
}

func sendServerInternalError(context *gin.Context, err error) {
    context.JSON(http.StatusInternalServerError, gin.H {
        "error": err.Error(),
    })
}

func ConvertFile(context *gin.Context, convertService *ConvertService) {
    var clientRequest ClientRequest

    if err := context.ShouldBind(&clientRequest); err != nil {
        slog.Error("missing or invalid parameters", "error", err)
        context.JSON(http.StatusBadRequest, gin.H {
            "error": "missing or invalid parameters: " + err.Error(),
        })
        return
    }
    // fmt.Printf("===== params: %+v\n", clientRequest)

    // TODO: validate params

    correlationID := uuid.NewString()
    requestResult, err := convertService.RequestConvert(&clientRequest, correlationID)
    if err != nil {
        slog.Error("convertService.FileUploadToMinIO() failed", "error", err)
        sendServerInternalError(context, err)
        return
    }
    // fmt.Printf("requestResult: %+v\n", requestResult)

    var convertResponse ConvertResponse // from ffmpegd
    var jobResponses []JobResponse // to web client
    for response := range convertService.ResponseConvert() {
        // fmt.Printf("Received a message: %s\n", response.Body)

        if err := json.Unmarshal(response.Body, &convertResponse); err != nil {
            slog.Error("json.Unmarshal() failed:", "error", err, "body", response.Body)
            sendServerInternalError(context, err)
            continue
        }

        // 요청한 JobID 면 처리하고, 맞지 않으면 재큐잉 한다.
        if _, exists := requestResult[convertResponse.JobID]; exists {
            jobID := convertResponse.JobID
            objectName := requestResult[jobID]

            convertedName := getConvertedFileName(&clientRequest, jobID)
            if convertedName == "" {
                convertedName = objectName
            }

            downloadURL, err := convertService.ObjectStorage.GetPresignedDownloadURL(objectName, convertedName, time.Hour * 6)
            if err != nil {
                slog.Error("convertService.ObjectStorage.GetPresignedDownloadURL() failed", "error", err)
                continue
            }
            
            if err := response.Ack(false); err != nil {
                slog.Error("response.Ack(false) failed", "error", err)
                continue
            }
            
            jobResponses = append(jobResponses, JobResponse {
                JobID: jobID,
                DownloadURL: downloadURL,
            })

            // 처리 완료된 요청 삭제
            delete(requestResult, jobID)

            // 모든 요청 처리시 아웃 루프
            if len(requestResult) == 0 {
                context.JSON(http.StatusOK, gin.H {
                    "results": jobResponses,
                })
                break
            }
        } else {
            err := response.Nack(false, true)
            if err != nil {
                slog.Error("response.Nack(false, true) failed", "error", err)
            }
        }
    }

}
