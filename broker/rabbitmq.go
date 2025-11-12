// Package broker
package broker

import (
    "encoding/json"
    "log/slog"
    "errors"
    amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitmqClient TODO: consider multiple channels and queues
type RabbitmqClient struct {
    conn *amqp.Connection
    ch *amqp.Channel
    queueName string
}

// NewRabbitmqClient TODO: consider multiple channels and queues
func NewRabbitmqClient(serverURL string, queueName string) (client *RabbitmqClient, err error) {
    var conn *amqp.Connection
    var ch *amqp.Channel

    conn, err = amqp.Dial(serverURL)
    if err != nil {
        slog.Error("amqp.Dial() is failed", "error", err)
        goto out
    }

    ch, err = conn.Channel()
    if err != nil {
        slog.Error("client.conn.Channel() is failed", "error", err)
        goto out
    }

    _, err = ch.QueueDeclare(
        queueName,
        true, // durable (메시지를 디스크에 저장하여 영속성 유지)
        false, // delete when unused (마지막 컨슈머 종료후 자동 삭제)
        false, // exclusive (현재의 단일 연결만 허용. 연결 종료후 자동 삭제)
        false, // nowait (서버의 큐 생성완료를 기다리지않고 리턴)
        nil,
        )
    if err != nil {
        slog.Error("client.ch.QueueDeclare() is failed", "error", err)
        goto out
    }

    return &RabbitmqClient {
        conn: conn,
        ch: ch,
        queueName: queueName,
    }, nil

    out:
    if conn != nil {
        if ch != nil {
            if err = ch.Close(); err != nil {
                slog.Error("client.ch.Close() is failed", "error", err)
            }
        }
        if err = conn.Close(); err != nil {
            slog.Error("client.conn.Close() is failed", "error", err)
        }
    }

    return nil, err
}

func (client *RabbitmqClient) Close() {
    if client.conn != nil {
        if client.ch != nil {
            if err := client.ch.Close(); err != nil {
                slog.Error("client.ch.Close() is failed", "error", err)
            }
        }
        if err := client.conn.Close(); err != nil {
            slog.Error("client.conn.Close() is failed", "error", err)
        }
    }
}

func (client *RabbitmqClient) Publish(data any, dataType string) (err error) {
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

    err = client.ch.Publish(
        "", // exchange (default 는 direct)
            // direct(routing key 와 정확히 매치된 큐에 전송)
            // fanout(routing key 무시하고 broadcasting)
            // topic(* 같은 패턴 적용, multicasting)
        client.queueName, // routing key
        false, // mandatory: 메시지가 어느큐에도 라우팅되지 않으면, Return 이벤트로 메시지 반환
        false, // immediate (deprecated): 즉시 소비 가능한 큐에 전달. 없을경우 반환.
        amqp.Publishing {
            ContentType: dataType,
            Body: body,
        },
        )
    if err != nil {
        slog.Error("client.ch.Publish() is failed", "error", err)
        return err
    }

    return nil
}

