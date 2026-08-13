package kafka

import "context"

type Event interface {
	EventType() string
	ResourceID() string
}

type EventPublisher struct {
	producer   Producer
	serializer Serializer
}

func NewEventPublisher(producer Producer, serializer Serializer) *EventPublisher {
	return &EventPublisher{
		producer:   producer,
		serializer: serializer,
	}
}

func (ep *EventPublisher) Publish(ctx context.Context, event Event) error {
	payload, err := ep.serializer.Serialize(event)
	if err != nil {
		return err
	}

	topic := "order-service." + event.EventType()
	key := []byte(event.ResourceID())

	if err := ep.producer.Publish(ctx, topic, key, payload); err != nil {
		return err
	}
	return nil
}
