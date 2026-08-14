package event

import (
	"fmt"
	"time"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateId() string
}

type BaseEvent struct {
	eventName   string
	occurredAt  time.Time
	aggregateId string
}

func NewBaseEvent(name, aggregateId string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventName:   name,
		occurredAt:  occurredAt,
		aggregateId: aggregateId,
	}
}

func (b BaseEvent) Validate() error {
	if b.eventName == "" {
		return fmt.Errorf("event name cannot be empty")
	}
	if b.aggregateId == "" {
		return fmt.Errorf("aggregate ID cannot be empty")
	}
	if b.occurredAt.IsZero() {
		return fmt.Errorf("occurred at time cannot be zero")
	}
	return nil
}

func (b BaseEvent) EventName() string     { return b.eventName }
func (b BaseEvent) OccurredAt() time.Time { return b.occurredAt }
func (b BaseEvent) AggregateId() string   { return b.aggregateId }
