// Package relay implementa o lado de leitura do Transactional Outbox
// (ver ADR-003): faz polling da tabela `outbox` de cada bounded context e
// publica os eventos não processados no Kafka, marcando-os como
// processados após a confirmação do broker.
package relay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	"github.com/google/uuid"
)

// OutboxReader é o subconjunto de shared/outbox.Repository que o relay
// precisa: ler eventos pendentes e marcá-los como processados. O lado de
// escrita (SaveEvents) pertence à aplicação de cada bounded context, não
// ao relay.
type OutboxReader interface {
	FetchUnprocessed(ctx context.Context, limit int) ([]outbox.Row, error)
	MarkProcessed(ctx context.Context, ids []uuid.UUID) error
}

// Publisher é satisfeito por shared/messaging/kafka.Producer. O relay
// publica o payload do outbox exatamente como foi persistido — não há
// nada para (re)serializar, o JSON já existe desde que a aplicação gravou
// o evento na mesma transação que o agregado (ver ADR-003, LEARNING-004).
type Publisher interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
}

// Source é um bounded context cuja tabela outbox o relay deve drenar. Name
// vira o prefixo do tópico Kafka (ex: "order-service"), seguindo a
// convenção da ADR-004: "<contexto>.<evento>".
type Source struct {
	Name   string
	Outbox OutboxReader
}

type Relay struct {
	sources   []Source
	publisher Publisher
	batchSize int
}

func New(publisher Publisher, batchSize int, sources ...Source) *Relay {
	return &Relay{
		sources:   sources,
		publisher: publisher,
		batchSize: batchSize,
	}
}

// RunOnce faz uma passada de polling em todas as sources. Uma falha em uma
// source, ou em uma linha específica, não impede o processamento das
// demais — cada linha que falhar simplesmente não é marcada como
// processada, e será tentada novamente na próxima passada. Retorna quantos
// eventos foram publicados com sucesso e o primeiro erro encontrado (se
// houver), para fins de observabilidade do chamador.
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	published := 0
	var firstErr error

	for _, source := range r.sources {
		rows, err := source.Outbox.FetchUnprocessed(ctx, r.batchSize)
		if err != nil {
			err = fmt.Errorf("relay: fetch unprocessed from %s: %w", source.Name, err)
			log.Println(err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		processed := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			topic := source.Name + "." + row.Type

			if err := r.publisher.Publish(ctx, topic, []byte(row.AggregateID), row.Payload); err != nil {
				err = fmt.Errorf("relay: publish outbox row %s (%s): %w", row.ID, topic, err)
				log.Println(err)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			processed = append(processed, row.ID)
			published++
		}

		if len(processed) == 0 {
			continue
		}

		if err := source.Outbox.MarkProcessed(ctx, processed); err != nil {
			err = fmt.Errorf("relay: mark processed for %s: %w", source.Name, err)
			log.Println(err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return published, firstErr
}

// Run chama RunOnce a cada `interval`, até que ctx seja cancelado.
func (r *Relay) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if published, err := r.RunOnce(ctx); published > 0 || err != nil {
				log.Printf("relay: cycle published=%d err=%v", published, err)
			}
		}
	}
}
