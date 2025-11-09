package broker

type Broker interface {
    Publish(data any, dataType string) error
    Close()
}
