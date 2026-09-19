# ADR-004 — Uso do Kafka

- Status: Accepted
- Date: 2026-09-18

## Context

O Hermes OrderFlow é composto por múltiplos bounded contexts
(order-service, inventory-service, e futuramente analytics-service,
notification-service) que precisam reagir a mudanças de estado uns dos
outros sem acoplamento direto entre si.

Exemplo motivador: quando um pedido é colocado (`order.placed`), o
inventory-service precisa reservar o estoque correspondente; quando a
reserva falha ou é bem-sucedida, o order-service precisa saber para
confirmar ou cancelar o pedido. Nenhum dos dois serviços deve chamar o
outro diretamente via HTTP síncrono para essa coordenação — isso criaria
acoplamento temporal (ambos precisam estar no ar ao mesmo tempo) e
dificultaria adicionar novos consumidores (ex: analytics-service) sem
alterar o produtor do evento.

## Decision

Foi escolhido o Kafka como broker de mensageria para comunicação
assíncrona entre bounded contexts, via arquitetura orientada a eventos
com coreografia (sem orquestrador central).

Convenções adotadas:

- Cada bounded context publica em tópicos prefixados com o próprio nome
  de serviço: `order-service.<evento>`, `inventory-service.<evento>`.
- O nome do evento (`EventName()`) segue o padrão `<contexto>.<fato>`
  (ex: `order.placed`, `stock.reserved`), consistente com o padrão de
  eventos de domínio descrito na ARCHITECTURE.md.
- A chave da mensagem é o `AggregateId()` do evento, garantindo que
  eventos do mesmo agregado sejam roteados para a mesma partição e
  processados em ordem.
- Nenhum serviço publica diretamente no Kafka a partir do caso de uso;
  a publicação passa pelo Transactional Outbox (ADR-003) e é feita pelo
  `cdc-connector`.

O cliente de produção (`Producer`, `Serializer`, `EventPublisher`) vive
em `shared/messaging/kafka`, reaproveitado por qualquer bounded context
que precise publicar eventos — ver ADR-005.

## Why Kafka

A escolha reflete tanto uma necessidade arquitetural quanto um objetivo
de aprendizado explícito do projeto (ver `docs/learning/`):

- Particionamento e ordenação por chave são um requisito real: eventos
  do mesmo agregado (ex: um mesmo SKU) precisam ser processados em
  ordem por qualquer consumidor.
- Retenção configurável permite que um consumidor novo (ex: um futuro
  analytics-service) se junte mais tarde e reprocesse o histórico de
  eventos, o que não seria possível com uma fila que descarta a
  mensagem após o consumo (ex: RabbitMQ em modo padrão).
- O modelo de consumer groups do Kafka se encaixa naturalmente na
  coreografia entre bounded contexts: múltiplos serviços podem consumir
  o mesmo tópico de forma independente.

## Alternatives Considered

### RabbitMQ

Mais simples de operar para filas ponto-a-ponto, mas o modelo de
exchanges/filas não oferece retenção e replay de eventos com a mesma
naturalidade que o Kafka, o que conflita com o objetivo de manter um
histórico de eventos de domínio reprocessável.

### Chamadas HTTP síncronas entre serviços

Reduziria a complexidade operacional inicial (sem broker), mas
acopla a disponibilidade dos serviços entre si e obriga o produtor do
evento a conhecer todos os consumidores interessados — o oposto do que
a arquitetura orientada a eventos busca.

## Trade-offs

### Benefícios

- Desacoplamento entre produtores e consumidores de eventos.
- Ordenação garantida por partição/chave.
- Retenção e replay de eventos.
- Escala horizontalmente via partições e consumer groups.

### Custos

- Complexidade operacional adicional (broker, Zookeeper/KRaft,
  monitoramento de consumer lag).
- Exige que consumidores sejam idempotentes, já que a entrega efetiva
  é at-least-once (consequência direta do Transactional Outbox, ver
  ADR-003).
- Serialização e evolução de schema dos eventos precisam de disciplina
  (hoje JSON simples, sem um registro de schema).

## Consequences

O Kafka roda localmente via `docker-compose.yml` (Zookeeper + um broker
único, `confluentinc/cp-kafka`). Em produção, esse broker single-node
não é adequado — a decisão de infraestrutura (Kafka self-hosted em
Kubernetes vs. MSK gerenciado) fica em aberto e deve ser revisitada
quando o deploy em nuvem for tratado.

Nenhum serviço ainda publica eventos de fato: o producer existe em
`shared/messaging/kafka`, mas só será exercitado quando o
`cdc-connector` for implementado (ver ADR-003).

## Security Considerations

A comunicação com o Kafka em `docker-compose.yml` usa `PLAINTEXT`, sem
autenticação nem TLS — aceitável apenas em ambiente de desenvolvimento
local. Em produção, o broker deve exigir TLS e autenticação
(SASL/SCRAM ou IAM, no caso de MSK), e os tópicos devem ter ACLs
restringindo quais serviços podem publicar e consumir de cada um.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-005 — Shared Kernel entre bounded contexts
