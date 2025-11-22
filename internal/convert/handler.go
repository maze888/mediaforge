// Package convert
package convert

import (
    "fmt"
    "log/slog"
    "net/http"
    "mime/multipart"
    "github.com/gin-gonic/gin"
)

// ClientRequest from web client
type ClientRequest struct {
    JobID []string `form:"JobID" binding:"required"`
    InputFormat []string `form:"InputFormat" binding:"required"`
    OutputFormat []string `form:"OutputFormat" binding:"required"`
    ConversionFiles []*multipart.FileHeader `form:"files" binding:"required"`
}

// ConvertRequest to ffmpegd
type ConvertRequest struct {
    JobID string
    CorrelationID string
    PresignedUploadURL string
    PresignedDownloadURL string
}

// ConvertResponse from ffmpegd
type ConvertResponse struct {
    JobID string `json:"job_id" binding:"required"`
    UploadedObjectName string `json:"uploaded_object_name" binding:"required"`
    ErrorMessage string `json:"error_message" binding:"required"`
    Status bool `json:"status" binding:"required"`
}
// 어디에 활용할지 아직 모르겠다.
// type ConvertResponseMap map[string]ConvertResponse // [JobID]ConvertResponse

func ConvertFile(context *gin.Context, convertService *ConvertService) {
    var params ClientRequest

    if err := context.ShouldBind(&params); err != nil {
        slog.Error("missing or invalid parameters", "error", err)
        context.JSON(http.StatusBadRequest, gin.H {
            "error": "missing or invalid parameters: " + err.Error(),
        })
        return
    }
    fmt.Printf("===== params: %+v\n", params)

    if err := convertService.FileUploadToMinio(&params); err != nil {
        slog.Error("convertService.FileUploadToMinIO() failed", "error", err)
        context.JSON(http.StatusInternalServerError, gin.H {
            "error": err.Error(),
        })
        return
    }
    
    context.JSON(http.StatusOK, gin.H {
        "message": "success",
    })
}
