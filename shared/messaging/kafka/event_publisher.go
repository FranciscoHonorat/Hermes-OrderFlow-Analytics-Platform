package kafka

import "context"

type Event interface {
	EventType() string
	ResourceID() string
}

// EventPublisher publica eventos de domínio no Kafka, prefixando o tópico
// com o nome do bounded context de origem (ex: "order-service",
// "inventory-service"). Cada serviço instancia o seu próprio publisher
// informando esse prefixo — o publisher em si é reutilizado por todos.
type EventPublisher struct {
	producer    Producer
	serializer  Serializer
	topicPrefix string
}

func NewEventPublisher(producer Producer, serializer Serializer, topicPrefix string) *EventPublisher {
	return &EventPublisher{
		producer:    producer,
		serializer:  serializer,
		topicPrefix: topicPrefix,
	}
}

func (ep *EventPublisher) Publish(ctx context.Context, event Event) error {
	payload, err := ep.serializer.Serialize(event)
	if err != nil {
		return err
	}

	topic := ep.topicPrefix + "." + event.EventType()
	key := []byte(event.ResourceID())

	if err := ep.producer.Publish(ctx, topic, key, payload); err != nil {
		return err
	}
	return nil
}
