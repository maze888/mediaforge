package broker
    
import amqp "github.com/rabbitmq/amqp091-go"

// Broker TODO: 공통 인터페이스로 개선이 가능한지... 현재는 RabbitMQ 구현체에 종속...
type Broker interface {
    Publish(data any, dataType, correlationID string) error
    Consume() (<-chan amqp.Delivery, error)
    Close()
}
