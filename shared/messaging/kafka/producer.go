package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	Client *kgo.Client
}

func NewProducer(brokers []string, clientID string) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID(clientID),
		// Sem isso, publicar no primeiro evento de um tópico que ainda não
		// existe falha com UNKNOWN_TOPIC_OR_PARTITION mesmo com
		// auto.create.topics.enable=true no broker — o cliente kgo não
		// pede criação automática por padrão.
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}

	return &Producer{
		Client: cl,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, key []byte, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}

	err := p.Client.ProduceSync(ctx, record).FirstErr()
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Close() error {
	p.Client.Close()
	return nil
}
