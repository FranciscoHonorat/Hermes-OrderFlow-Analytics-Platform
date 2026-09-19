# ADR-006 — cdc-connector como relay de polling do Outbox

- Status: Accepted
- Date: 2026-09-19

## Context

Desde a ADR-003, order-service e inventory-service gravam eventos de
domínio na própria tabela `outbox`, dentro da mesma transação que
persiste o agregado. Mas até este ponto do roadmap (Fase 0 e Fase 1),
nada lia essa tabela: `shared/outbox.Repository` tinha `SaveEvents`
funcionando, e `FetchUnprocessed`/`MarkProcessed` existiam como contrato
sem nenhum processo real os chamando. Os eventos ficavam presos.

O `cdc-connector` é esse processo. Ele precisa:

- ler a tabela `outbox` de múltiplos bounded contexts (hoje order-service
  e inventory-service, mais no futuro);
- publicar cada evento não processado no Kafka, no tópico
  `<bounded-context>.<evento>` (ADR-004);
- marcar a linha como processada só depois da confirmação do broker;
- continuar funcionando mesmo que um bounded context específico esteja
  temporariamente inacessível.

## Decision

O `cdc-connector` v1 é um relay por **polling**, não CDC baseado em log
de replicação (WAL/Debezium) — essa evolução já estava prevista como
"stretch goal" na ADR-003 e continua em aberto.

Design:

- Um único processo Go, sem HTTP de negócio (só `/health`), que mantém
  uma goroutine de loop (`relay.Relay.Run`) rodando a cada
  `POLL_INTERVAL` (padrão 2s).
- Cada bounded context é uma `Source` configurada via variável de
  ambiente `SOURCES` no formato `nome:banco,nome:banco` (ex:
  `order-service:orderflow,inventory-service:inventory`) — todas as
  sources compartilham host/usuário/senha do Postgres (ambiente de
  desenvolvimento local), só o nome do banco varia.
- Para cada source, a cada ciclo: `FetchUnprocessed(ctx, batchSize)` →
  publica cada linha via `shared/messaging/kafka.Producer.Publish` →
  `MarkProcessed` apenas para as linhas publicadas com sucesso.
- Uma falha em uma source, ou em uma linha específica, não interrompe o
  processamento das demais — cada linha que falhar simplesmente não é
  marcada como processada, e será tentada de novo no próximo ciclo
  (idempotência fica por conta do consumidor, ver ADR-003/ADR-004).
- O relay em si (`internal/relay`) não conhece Postgres nem Kafka
  diretamente — depende de duas interfaces pequenas (`OutboxReader`,
  `Publisher`), o que permite testá-lo inteiramente com fakes (ver
  `relay_test.go`) e separadamente com Postgres e Kafka reais via
  testcontainers (`integration_test.go`).

## Why polling (e não CDC baseado em log)

Debezium/WAL logical replication captura qualquer mudança na tabela de
origem sem exigir que a aplicação grave explicitamente na tabela
outbox, e evita o custo de polling repetido no banco. Foi descartado
para a v1 pelo mesmo motivo já registrado na ADR-003: adiciona
complexidade operacional (Kafka Connect, conectores Debezium,
gerenciamento de replication slots) que não se justifica no estágio
atual do projeto. Polling sobre uma tabela outbox pequena e indexada
(`idx_outbox_unprocessed`, ver migrations) é barato o suficiente para o
volume atual, e o contrato (`FetchUnprocessed`/`MarkProcessed`) foi
desenhado para que trocar a implementação por CDC baseado em log no
futuro não exija mudar a interface que o resto do sistema depende.

## Alternatives Considered

### Um relay por bounded context (em vez de um processo central)

Cada serviço poderia ter seu próprio processo de relay, acoplado a ele.
Isso removeria a necessidade de um processo com credenciais de leitura
em múltiplos bancos (ver Security Considerations), mas duplicaria o
mesmo loop de polling N vezes e tornaria mais difícil coordenar
mudanças na lógica de entrega (retry, backoff, observabilidade) — hoje
concentradas em um único lugar.

### Publicar diretamente da aplicação, sem outbox nem relay

Já descartado na ADR-003: reintroduz a janela de inconsistência entre
persistir o agregado e publicar o evento.

## Trade-offs

### Benefícios

- Fecha o ciclo completo do Transactional Outbox: eventos gravados
  hoje finalmente chegam ao Kafka.
- Reaproveita inteiramente o shared kernel (ADR-005): nenhuma lógica de
  outbox ou Kafka é duplicada.
- Fácil de testar: lógica de negócio do relay isolada de I/O real via
  interfaces pequenas.
- Adicionar um novo bounded context como source é uma linha na
  variável `SOURCES`, sem mudar código.

### Custos

- Latência de entrega limitada por `POLL_INTERVAL` (hoje 2s) — não é
  near-real-time.
- Um processo central com acesso de leitura/escrita à tabela outbox de
  todo bounded context é um ponto único de acoplamento operacional: se
  o `cdc-connector` cair, nenhum evento de nenhum serviço é entregue
  (mas nenhum dado é perdido — a tabela outbox continua acumulando).
- Custo de polling no Postgres de cada source a cada ciclo, mesmo
  quando não há eventos novos.

## Consequences

    order-service (DB orderflow)  ──┐
                                     ├─▶ cdc-connector ──▶ Kafka
    inventory-service (DB inventory)┘        │
                                       FetchUnprocessed
                                       Publish
                                       MarkProcessed

Cada bounded context continua sem saber que o `cdc-connector` existe —
ele só grava na própria tabela outbox (ADR-003). O `cdc-connector` é o
único processo do sistema com uma conexão Postgres por bounded context
simultaneamente.

## Security Considerations

O `cdc-connector` precisa de credenciais válidas para o banco de dados
de todo bounded context que ele drena — uma exceção deliberada ao
princípio de que cada serviço é dono exclusivo do seu banco (ADR-002).
Em produção, o usuário de banco usado pelo `cdc-connector` deve ter
permissão restrita a `SELECT`/`UPDATE` apenas na tabela `outbox` de
cada banco, nunca acesso amplo às tabelas de domínio (`orders`,
`stock_items`, etc.).

Durante a implementação desta ADR também foram corrigidos dois
problemas de configuração do Kafka local que bloqueavam qualquer
publicação real (ver LEARNING-005): o broker do `docker-compose.yml`
estava em crash-loop por um conflito de listeners, e o
`kafka.Producer` não solicitava criação automática de tópico. Nenhum
dos dois é específico do `cdc-connector`, mas só foram descobertos
porque ele foi o primeiro processo do sistema a de fato tentar publicar
no Kafka.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-004 — Uso de Kafka
- ADR-005 — Shared Kernel entre bounded contexts
