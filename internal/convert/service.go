// Package convert
package convert

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "mediaforge2/internal/infra/queue"
    "mediaforge2/internal/infra/storage"
    "mediaforge2/internal/util"
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

func (cs *ConvertService) FileUploadToMinio(clientRequest *ClientRequest) error {
    for _, file := range clientRequest.ConversionFiles {
        // 객체명 중복 방지
        uniqueObjectName := fmt.Sprintf("%s%s", uuid.NewString(), util.GetFileExt(file.Filename))
        if err := cs.ObjectStorage.Upload(file, uniqueObjectName); err != nil {
            return fmt.Errorf("cs.objectStorage.Upload")
        }
    } 

    return nil
}
