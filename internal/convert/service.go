package convert

import (
    "fmt"
    "errors"
    "time"
    "net/http"
    "log/slog"

    "mediaforge/storage"
    "mediaforge/broker"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type ConvertService struct {
    storage storage.ObjectStorage
    broker broker.Broker
}

type FormatRequest struct {
    InputFormat, OutputFormat string
}

type ConvertRequest struct {
    InputFormat, OutputFormat string
    PresignedUploadURL, PresignedDownloadURL string
}

func newConvertService(bucketName, brokerServerURL, brokerQueueName string) (*ConvertService, error) {
    storage, err := storage.NewMinioStorage(bucketName); 
    if err != nil {
        slog.Error("Failed to create MinIO storage", "error", err)
        return nil, err
    }

    client, err := broker.NewRabbitmqClient(brokerServerURL, brokerQueueName)
    if err != nil {
        slog.Error("broker.NewRabbitmqClient() is failed")
        return nil, err
    }

    return &ConvertService {
        storage: storage,
        broker: client,
    }, nil
}

// Upload TODO: 여러 파일 동시 처리(현재는 파일 하나만 처리됨)
func (service *ConvertService) Upload(context *gin.Context, params *FormatRequest) (req *ConvertRequest, err error) {
    var uploadURL, downloadURL string

    uploadFileNames, err := service.upload(context)
    if err != nil {
        slog.Error("upload() is failed", "error", err)
        context.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
        return nil, err
    }

    for _, uploadFileName := range uploadFileNames {
        // ffmpegd 처리끝난 결과물 올릴 URL
        uploadURL2, err := service.storage.GetPresignedUploadURL(fmt.Sprintf("%s.%s", uploadFileName, params.OutputFormat), time.Minute * 5)
        if err != nil {
            slog.Error("GetPresignedUploadURL() is failed", "error", err)
            context.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
            return nil, err
        }

        // ffmpegd 처리할 결과물 받을 URL
        downloadURL2, err := service.storage.GetPresignedDownloadURL(uploadFileName, time.Minute * 5)
        if err != nil {
            slog.Error("GetPresignedUploadURL() is failed", "error", err)
            context.JSON(http.StatusInternalServerError, gin.H{ "error": err, })
            return nil, err
        }

        // 현재는 파일 하나만 처리함...
        uploadURL = uploadURL2.String()
        downloadURL = downloadURL2.String()
        break
    }

    // TODO: rabbitmq 를 통해 ffmpegd 로부터 처리 완료 메시지 수신

    // TODO: 웹클라이언트에게 처리 결과 전송 (download url 포함)
    //       downloadURL 을 다시 생성하되, 옵션 추가로 
    //       다운로드된 파일명은 uuid 제거된 원래 파일명에
    //       마지막 확장자만 .mp3 붙이도록 파일명 지정한다.
    return &ConvertRequest {
        InputFormat: params.InputFormat,
        OutputFormat: params.OutputFormat,
        PresignedUploadURL: uploadURL,
        PresignedDownloadURL: downloadURL,
    } , nil
}

func (service *ConvertService) Request(convertRequest *ConvertRequest) (error) {
    return service.broker.Publish(convertRequest, "application/json")
}

func (service *ConvertService) upload(context *gin.Context) (uploadFileNames []string, err error) {
    form, err := context.MultipartForm()
    if err != nil {
        context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return nil, err
    }

    files := form.File["files"]
    if len(files) == 0 {
        context.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
        return nil, errors.New("no file uploaded")
    }

    for _, file := range files {
        src, err := file.Open()
        if err != nil {
            context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            slog.Error("File open is failed", "error", err)
            return nil, err
        }

        // 파일 중복 업로드 방지를 위한 UUID 추가
        uploadFileName := file.Filename + "-" + uuid.New().String()
        if err := service.storage.Upload(src, uploadFileName, file.Size); err != nil {
            context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            slog.Error("File open is failed", "error", err)
            src.Close()
            return nil, err
        }
        src.Close()
        uploadFileNames = append(uploadFileNames, uploadFileName)
    }

    context.JSON(200, gin.H {
        "message": fmt.Sprintf("%d file(s) uploaded successfully", len(files)),
    })

    return uploadFileNames, nil
}

