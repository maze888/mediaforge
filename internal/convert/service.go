// Package convert
package convert

import (
    "fmt"
    "time"
    "net/http"

    "mediaforge/internal/infra/queue"
    "mediaforge/internal/infra/storage"
    
    "github.com/gin-gonic/gin"
    amqp "github.com/rabbitmq/amqp091-go"
)

// const (
//     MethodGet     = "GET"
//     MethodHead    = "HEAD"
//     MethodPost    = "POST"
//     MethodPut     = "PUT"
//     MethodPatch   = "PATCH"
//     MethodDelete  = "DELETE"
//     MethodConnect = "CONNECT"
//     MethodOptions = "OPTIONS"
//     MethodTrace   = "TRACE"
// )

type ConvertService struct {
    Router *gin.Engine
    MessageQueue *queue.RabbitmqClient
    ObjectStorage *storage.MinioClient
}

func NewConvertService(r *gin.Engine, q *queue.RabbitmqClient, s *storage.MinioClient) *ConvertService {
    return &ConvertService {
        Router: r,
        MessageQueue: q,
        ObjectStorage: s,
    }
}

func (cs *ConvertService) AddHandler(httpMethod, resourcePath string, handleFunc func(context *gin.Context, convertService *ConvertService)) error {
    switch httpMethod {
    case http.MethodPost:
        cs.Router.POST(resourcePath, func(context *gin.Context) {
            handleFunc(context, cs)
        })
    default:
        return fmt.Errorf("invalid method type: %v", httpMethod)
    }

    return nil
}

func (cs *ConvertService) RequestConvert(clientRequest *ClientRequest, correlationID string) (map[string]string, error) {
    requestResult := make(map[string]string)

    for i, file := range clientRequest.ConversionFiles {
        originalObjectName := fmt.Sprintf("%s.%s", clientRequest.JobID[i], clientRequest.InputFormat[i])
        // File Upload To Minio
        if err := cs.ObjectStorage.Upload(file, originalObjectName); err != nil {
            return nil, fmt.Errorf("cs.objectStorage.Upload() failed: %w", err)
        }

        convertedObjectName := fmt.Sprintf("%s.%s", clientRequest.JobID[i], clientRequest.OutputFormat[i])
        // To ffmpegd
        presignedUploadURL, err := cs.ObjectStorage.GetPresignedUploadURL(convertedObjectName, time.Hour * 6)
        if err != nil {
            return nil, fmt.Errorf("cs.objectStorage.GetPresignedUploadURL() failed: %w", err)
        }
        presignedDownloadURL, err := cs.ObjectStorage.GetPresignedDownloadURL(originalObjectName, file.Filename, time.Hour * 6) 
        if err != nil {
            return nil, fmt.Errorf("cs.objectStorage.GetPresignedUploadURL() failed: %w", err)
        }

        // Send To Request MessageQueue
        convertRequest := ConvertRequest {
            JobID: clientRequest.JobID[i],
            FileName: file.Filename,
            InputFormat: clientRequest.InputFormat[i],
            OutputFormat: clientRequest.OutputFormat[i],
            PresignedUploadURL: presignedUploadURL,
            PresignedDownloadURL: presignedDownloadURL,
        }
        if err := cs.MessageQueue.Publish(&convertRequest, "application/json", correlationID); err != nil {
            return nil, fmt.Errorf("cs.MessageQueue.Publish() failed: %w", err)
        }

        // ffmpegd 가 작업 완료후 MinIO 에 올린 오브젝트명
        requestResult[clientRequest.JobID[i]] = convertedObjectName
    } 

    return requestResult, nil
}

func (cs *ConvertService) ResponseConvert() (<-chan amqp.Delivery){
    return cs.MessageQueue.Consume()
}
