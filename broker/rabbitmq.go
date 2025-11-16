// Package broker
package broker

import (
    "time"
    "log/slog"
    "errors"
    "encoding/json"

    "github.com/google/uuid"

    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitmqClient struct {
    conn *amqp.Connection
    requestChannel, responseChannel *amqp.Channel
    requestQueue, responseQueue amqp.Queue
    consumerChannel <-chan amqp.Delivery
}

func NewRabbitmqClient(serverURL string, requestQueueName, responseQueueName string) (client *RabbitmqClient, err error) {
    client = &RabbitmqClient {}

    client.conn, err = amqp.Dial(serverURL)
    if err != nil {
        slog.Error("amqp.Dial() is failed", "error: ", err)
        goto out
    }

    client.requestChannel, err = client.conn.Channel()
    if err != nil {
        slog.Error("client.conn.Channel(REQUEST) is failed", "error: ", err)
        goto out
    }
    
    client.responseChannel, err = client.conn.Channel()
    if err != nil {
        slog.Error("client.conn.Channel(RESPONSE) is failed", "error: ", err)
        goto out
    }

    // 대규모 처리시 문제 있음
    // if err = client.responseChannel.Qos(1, 0, false); err != nil {
    //     slog.Error("client.responseChannel.Qos() is failed", "error: ", err)
    //     goto out
    // }
    
    client.requestQueue, err = client.requestChannel.QueueDeclare(
        requestQueueName,
        true, // durable (메시지를 디스크에 저장하여 영속성 유지)
        false, // delete when unused (마지막 컨슈머 종료후 자동 삭제)
        false, // exclusive (현재의 단일 연결만 허용. 연결 종료후 자동 삭제)
        false, // nowait (서버의 큐 생성완료를 기다리지않고 리턴)
        nil,
        )
    if err != nil {
        slog.Error("client.requestChannel.QueueDeclare(REQUEST) is failed", "error: ", err)
        goto out
    }
    
    client.responseQueue, err = client.responseChannel.QueueDeclare(
        responseQueueName,
        true, // durable (메시지를 디스크에 저장하여 영속성 유지)
        false, // delete when unused (마지막 컨슈머 종료후 자동 삭제)
        false, // exclusive (현재의 단일 연결만 허용. 연결 종료후 자동 삭제)
        false, // nowait (서버의 큐 생성완료를 기다리지않고 리턴)
        nil,
        )
    if err != nil {
        slog.Error("client.responseChannel.QueueDeclare(RESPONSE) is failed", "error: ", err)
        goto out
    }
    
    client.consumerChannel, err = client.responseChannel.Consume(
		client.responseQueue.Name,
		"",
		false, // autoAck = false → 수동 ack
		false,
        false,
		false,
		nil,
	)
    if err != nil {
        slog.Error("client.responseChannel.Consume() is failed", "error: ", err)
        goto out
    }


    return

    out:
    client.Close()

    return
}

func (client *RabbitmqClient) Close() {
    if client.conn != nil {
        if client.requestChannel != nil {
            if err := client.requestChannel.Close(); err != nil {
                slog.Error("client.requestChannel.Close() is failed", "error", err)
            }
        }
        if client.responseChannel != nil {
            if err := client.responseChannel.Close(); err != nil {
                slog.Error("client.responseChannel.Close() is failed", "error", err)
            }
        }
        if err := client.conn.Close(); err != nil {
            slog.Error("client.conn.Close() is failed", "error", err)
        }
    }
}

func (client *RabbitmqClient) Publish(data any, dataType, correlationID string) (err error) {
    var body []byte

    switch dataType {
    case "text/plain":
        // TODO: implement
        break
    case "application/json":
        body, err = json.Marshal(data)
        if err != nil {
            slog.Error("json.Marshal() is failed", "error", err)
            return err
        }
    case "application/octet-stream":
        // TODO: implement
        break
    default:
        return errors.New("invalid data type")
    }

    err = client.requestChannel.Publish(
        "", // exchange (default 는 direct)
            // direct(routing key 와 정확히 매치된 큐에 전송)
            // fanout(routing key 무시하고 broadcasting)
            // topic(* 같은 패턴 적용, multicasting)
        client.requestQueue.Name, // routing key
        false, // mandatory: 메시지가 어느큐에도 라우팅되지 않으면, Return 이벤트로 메시지 반환
        false, // immediate (deprecated): 즉시 소비 가능한 큐에 전달. 없을경우 반환.
        newPublishMessage(body, dataType, client.responseQueue.Name, correlationID),
        )
    if err != nil {
        slog.Error("client.requestChannel.Publish() is failed", "error", err)
        return err
    }

    return nil
}

func (client *RabbitmqClient) Consume() (<-chan amqp.Delivery) {
    return client.consumerChannel
}

func newPublishMessage(data []byte, dataType, replyTo, correlationID string) amqp.Publishing {
    return amqp.Publishing {
        ContentType: dataType,
        Body: data,
        MessageId: uuid.NewString(),
        CorrelationId: correlationID,
        ReplyTo: replyTo,
        Timestamp: time.Now(),
    }
}


