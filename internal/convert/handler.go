// Package convert
package convert

import (
    "github.com/gin-gonic/gin"
    "fmt"
    "log/slog"
    "net/http"
    "errors"
    "time"
    "mediaforge/storage"
    "github.com/google/uuid"
)

type FormatRequest struct {
    InputFormat string `form:"inputFormat" json:"inputformat" binding:"required"`
    OutputFormat string `form:"outputFormat" json:"outputformat" binding:"required"`
}

func AddHandler(router *gin.Engine, minioStorage *storage.MinioStorage) {
    router.POST("/convert", func(ctx *gin.Context) {
        convert(ctx, minioStorage)
    })
}

func convert(ctx *gin.Context, minioStorage *storage.MinioStorage) {
    var params FormatRequest

    if err := ctx.ShouldBind(&params); err != nil {
        slog.Info("missing or invalid parameters", "error", err)
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "missing or invalid parameters: " + err.Error(),
        })
        return
    }

    // TODO: format parameter validate

    // minio 로 파일 업로드
    uploadFileNames, err := upload(ctx, minioStorage, &params)
    if err != nil {
        slog.Error("upload() is failed", "error", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
        return
    }

    for _, uploadFileName := range uploadFileNames {
        // ffmpegd 처리끝난 결과물 올릴 URL
        uploadURL, err := minioStorage.GetPresignedUploadURL(fmt.Sprintf("%s.%s", uploadFileName, params.OutputFormat), time.Minute * 5)
        if err != nil {
            slog.Error("GetPresignedUploadURL() is failed", "error", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
            return
        }
        
        // ffmpegd 처리할 결과물 받을 URL
        downloadURL, err := minioStorage.GetPresignedDownloadURL(fmt.Sprintf("%s.%s", uploadFileName, params.InputFormat), time.Minute * 5)
        if err != nil {
            slog.Error("GetPresignedUploadURL() is failed", "error", err)
            ctx.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
            return
        }

        fmt.Printf("uploadURL: %s\n", uploadURL)
        fmt.Printf("downloadURL: %s\n", downloadURL)
    
        // TODO: rabbitmq 를 통해 ffmpegd 에게 처리 요청 메시지 송신
    }

    // TODO: rabbitmq 를 통해 ffmpegd 로부터 처리 완료 메시지 수신
    
    // TODO: 웹클라이언트에게 처리 결과 전송 (download url 포함)
}

func makePresignedDownloadURLs(uploadFileNames []string) (presignedDownloadURLs []string, err error) {
    return presignedDownloadURLs, err 
}

func upload(ctx *gin.Context, minioStorage *storage.MinioStorage, params *FormatRequest) (uploadFileNames []string, err error) {
    form, err := ctx.MultipartForm()
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return nil, err
    }

    files := form.File["files"]
    if len(files) == 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
        return nil, errors.New("no file uploaded")
    }

    for _, file := range files {
        src, err := file.Open()
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            slog.Error("File open is failed", "error", err)
            return nil, err
        }
        defer src.Close()

        // 파일 중복 업로드 방지를 위한 UUID 추가
        uploadFileName := fmt.Sprintf("%s-%s", file.Filename, uuid.New().String(), params.InputFormat)
        if err := minioStorage.Upload(src, fmt.Sprintf("%s.%s", uploadFileName, params.InputFormat), file.Size); err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            slog.Error("File open is failed", "error", err)
            return nil, err
        }
        uploadFileNames = append(uploadFileNames, uploadFileName)
    }

    ctx.JSON(200, gin.H {
        "message": fmt.Sprintf("%d file(s) uploaded successfully", len(files)),
    })

    return uploadFileNames, nil
}
