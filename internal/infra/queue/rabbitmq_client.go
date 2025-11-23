// Package queue
package queue

import (
    "fmt"
    "encoding/json"

    "github.com/google/uuid"
    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqClient struct {
    conn *amqp.Connection
    commandChannel *amqp.Channel
    requestQueue, responseQueue amqp.Queue
    consumerChannel <-chan amqp.Delivery
}

func NewRabbitmqClient(url string, reqQName, resQName string) (client *RabbitmqClient, err error) {
    client = &RabbitmqClient {}

    client.conn, err = amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("amqp.Dial() failed: %w", err)
    }

    client.commandChannel, err = client.conn.Channel()
    if err != nil {
        return nil, fmt.Errorf("client.conn.Channel() failed: %w", err)
    }

    client.requestQueue, err = client.commandChannel.QueueDeclare(
        reqQName,
        true,  // durable (메시지를 디스크에 저장하여 영속성 유지)
        false, // delete when unused (마지막 컨슈머 종료후 자동 삭제)
        false, // exclusive (현재의 단일 연결만 허용. 연결 종료후 자동 삭제)
        false, // nowait (서버의 큐 생성완료를 기다리지않고 리턴)
        nil,
        ) 
    if err != nil {
        return nil, fmt.Errorf("client.commandChannel.QueueDeclare(REQUEST) failed: %w", err)
    }

    client.responseQueue, err = client.commandChannel.QueueDeclare(
        resQName,
        true, 
        false,
        false,
        false,
        nil,
        ) 
    if err != nil {
        return nil, fmt.Errorf("client.commandChannel.QueueDeclare(RESPONSE) failed: %w", err)
    }

    client.consumerChannel, err = client.commandChannel.Consume(
        client.responseQueue.Name,
        "",
        false, // autoAck
        false,
        false,
        false,
        nil,
        )
    if err != nil {
        return nil, fmt.Errorf("client.commandChannel.Consume() failed: %w", err)
    }

    return client, err
}

func (client *RabbitmqClient) Publish(data any, toDataType, correlationID string) (err error) {
    var body []byte

    switch toDataType {
    case "application/json":
        body, err = json.Marshal(data)
        if err != nil {
            return fmt.Errorf("json.Marshal() failed: %w", err)
        }
    default:
        // TODO: 에러 처리 Error.Is 개선(?)
        return fmt.Errorf("invalid data type: %v", toDataType)
    }

    err = client.commandChannel.Publish(
        "", // exchange (default: direct)
            // direct: 정확히 매치된 라우팅키 큐에 전송
            // fanout: 라우팅키 무시하고 브로드캐스팅
            // topic:  * 같은 패턴 적용, 멀티캐스팅)
        client.requestQueue.Name, // 라우팅키
        false, // mandatory: 라우팅 되지 않읗시, 리턴 이벤트로 메시지 반환
        false, // immediate (deprecated)
        client.newPublishMessage(body, toDataType, correlationID),
        )
    if err != nil {
        return fmt.Errorf("client.commandChannel.Publish() failed: %w", err)
    }

    return err
}

func (client *RabbitmqClient) newPublishMessage(data []byte, dataType, correlationID string) amqp.Publishing {
    return amqp.Publishing {
        ContentType: dataType,
        Body: data,
        MessageId: uuid.NewString(),
        CorrelationId: correlationID,
        ReplyTo: client.responseQueue.Name,
    }
}

func (client *RabbitmqClient) Consume() (<-chan amqp.Delivery) {
    return client.consumerChannel
}

