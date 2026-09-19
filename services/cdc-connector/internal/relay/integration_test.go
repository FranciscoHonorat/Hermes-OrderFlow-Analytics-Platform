//go:build integration

package relay_test

import (
	"context"
	"testing"
	"time"

	"github.com/FranciscoHonorat/ordemflow/services/cdc-connector/internal/relay"
	"github.com/FranciscoHonorat/ordemflow/shared/messaging/kafka"
	"github.com/FranciscoHonorat/ordemflow/shared/outbox"
	sharedpg "github.com/FranciscoHonorat/ordemflow/shared/postgres"

	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
)

// A tabela outbox tem o mesmo schema em todo bounded context (ver
// migrations/000002_create_outbox.up.sql do order-service e do
// inventory-service) — o teste recria essa tabela diretamente em vez de
// depender do diretório de migrations de outro módulo, já que o
// cdc-connector não é dono de nenhum schema próprio.
const outboxDDL = `
CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    aggregate_id TEXT NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ
);
`

func setupPostgres(t *testing.T) *sharedpg.DB {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("cdc_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sharedpg.NewConnection(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(db.Close)

	_, err = db.Pool.Exec(ctx, outboxDDL)
	require.NoError(t, err)

	return db
}

func setupKafka(t *testing.T) []string {
	t.Helper()
	ctx := context.Background()

	// confluentinc/confluent-local é a imagem que o próprio módulo
	// testcontainers sabe configurar em modo KRaft de nó único (gera o
	// CLUSTER_ID sozinha); é diferente da imagem confluentinc/cp-kafka
	// usada no docker-compose.yml (que roda com Zookeeper) — aqui o único
	// objetivo é ter um broker Kafka real e efêmero para o teste.
	container, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.6.0")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	brokers, err := container.Brokers(ctx)
	require.NoError(t, err)

	return brokers
}

// TestRelay_PublishesOutboxRowsToKafka é a prova de ponta a ponta do
// Fase 2 do roadmap: um evento gravado na tabela outbox (como a aplicação
// já faz hoje, dentro da mesma transação do agregado — ADR-003) chega de
// fato ao Kafka depois de uma passada do relay, e a linha é marcada como
// processada.
func TestRelay_PublishesOutboxRowsToKafka(t *testing.T) {
	db := setupPostgres(t)
	brokers := setupKafka(t)
	ctx := context.Background()

	rowID := uuid.New()
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO outbox (id, aggregate_id, type, payload, created_at) VALUES ($1, $2, $3, $4, now())`,
		rowID, "order-123", "order.placed", []byte(`{"hello":"world"}`),
	)
	require.NoError(t, err)

	outboxRepo := outbox.NewPostgresRepository(db.Pool)

	producer, err := kafka.NewProducer(brokers, "cdc-connector-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = producer.Close() })

	r := relay.New(producer, 10, relay.Source{Name: "order-service", Outbox: outboxRepo})

	published, err := r.RunOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, published)

	remaining, err := outboxRepo.FetchUnprocessed(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, remaining, "the row must be marked processed after a successful publish")

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics("order-service.order.placed"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	require.NoError(t, err)
	defer consumer.Close()

	fetchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	fetches := consumer.PollFetches(fetchCtx)
	require.Empty(t, fetches.Errors())

	records := fetches.Records()
	require.Len(t, records, 1)
	require.Equal(t, "order-123", string(records[0].Key))
	require.JSONEq(t, `{"hello":"world"}`, string(records[0].Value))
}
