# ADR-005 — Shared Kernel entre bounded contexts

- Status: Accepted
- Date: 2026-09-18

## Context

Até este ponto, `BaseEvent`/`DomainEvent`, o cliente Kafka
(`Producer`/`Serializer`/`EventPublisher`) e a implementação do
Transactional Outbox (ADR-003) viviam inteiramente dentro do
order-service.

Ao iniciar a implementação do inventory-service como um segundo bounded
context que também precisa emitir eventos de domínio e gravá-los em uma
tabela outbox, duas opções surgiram:

1. Duplicar esse código dentro de `services/inventory-service`, como já
   havia começado a acontecer (os erros de domínio do inventory-service
   já viviam soltos em `internal/domain/stock/errStock.go`, sem seguir
   o padrão `domain-errors` do order-service).
2. Extrair a parte desse código que não é específica de nenhum bounded
   context para um módulo compartilhado.

Duplicar o publisher Kafka, por exemplo, também exporia um bug latente:
o tópico era montado com o prefixo `"order-service."` fixo no código
(`infrastructure/messaging/kafka/event_publisher.go`), o que produziria
tópicos incorretos (`order-service.stock.reserved`) se copiado
ingenuamente para o inventory-service.

## Decision

Foi criado um Shared Kernel (no sentido de DDD — um submódulo
explicitamente compartilhado entre bounded contexts, com mudanças
coordenadas) em `shared/`, hoje com três pacotes:

- `shared/events`: a interface `DomainEvent` e o struct `BaseEvent` que
  todo evento de domínio de todo bounded context deve compor, conforme
  já determinava a ARCHITECTURE.md (seção 4) antes mesmo de existir um
  segundo serviço para reaproveitá-lo.
- `shared/messaging/kafka`: `Producer`, `Serializer` e `EventPublisher`.
  O `EventPublisher` passou a receber o prefixo de tópico como
  parâmetro do construtor em vez de tê-lo fixo no código — correção
  necessária para o reuso entre serviços.
- `shared/outbox`: o contrato `Repository` (`SaveEvents`,
  `FetchUnprocessed`, `MarkProcessed`) e sua implementação em
  PostgreSQL, usados por qualquer bounded context que adote o
  Transactional Outbox (ADR-003).

Cada bounded context continua definindo sua própria porta
`output.OutboxRepository` em `application/port/output`, mas como um
alias de tipo para `shared/outbox.Repository` — o ponto de extensão
arquitetural (a porta hexagonal) continua existindo em cada serviço,
apenas sua definição concreta é compartilhada.

O que **não** foi movido para o shared kernel: agregados, value objects
de negócio (SKU, Money, OrderID), unit of work e qualquer coisa
específica do domínio de um bounded context. Isso preserva o princípio
da ARCHITECTURE.md de que cada bounded context é independente — o shared
kernel existe apenas para infraestrutura de plataforma (como um evento
chega ao Kafka), nunca para regras de negócio.

## Why a Shared Kernel

A alternativa de duplicar esse código em cada serviço pareceria, à
primeira vista, mais alinhada ao isolamento entre bounded contexts que
a ARCHITECTURE.md defende. Mas o que estava sendo duplicado não era
regra de negócio — era a mecânica de "como registrar que um evento
aconteceu" e "como esse evento chega ao Kafka", que é idêntica em todo
o sistema por construção (todos usam Transactional Outbox e Kafka,
ADR-003 e ADR-004).

Duplicar esse código teria dois custos concretos:

- Divergência silenciosa: um bug corrigido no producer do
  order-service (como o prefixo de tópico fixo, corrigido nesta mesma
  mudança) não se propagaria automaticamente para o inventory-service.
- Atrito para o `cdc-connector`: ele precisa ler a tabela outbox de
  todo bounded context da mesma forma. Sem um contrato compartilhado,
  cada serviço poderia divergir na forma como marca eventos como
  processados.

## Alternatives Considered

### Duplicar o código em cada serviço

Mantém cada bounded context fisicamente isolado (sem dependência de
módulo entre eles), ao custo de divergência entre implementações que
deveriam ser idênticas, e re-trabalho a cada novo serviço.

### Publicar `shared` como um módulo Go versionado e importado via proxy

Formalizaria melhor a fronteira do shared kernel (com versionamento
semântico e changelog), mas adiciona um processo de release que não se
justifica no estágio atual do projeto. Hoje `shared` é resolvido
localmente via workspace (`go.work`) e `replace` nos `go.mod` de cada
serviço, sem publicação.

## Trade-offs

### Benefícios

- Elimina duplicação de código que não é decisão de negócio de nenhum
  bounded context.
- Uma correção de bug na mecânica de outbox/Kafka beneficia todos os
  serviços automaticamente.
- Reduz o trabalho de scaffolding de cada novo bounded context.

### Custos

- Introduz uma dependência de compilação entre todos os serviços e
  `shared`: uma mudança incompatível em `shared/events.DomainEvent`,
  por exemplo, quebra todo bounded context de uma vez.
- Exige disciplina para não deixar regra de negócio "vazar" para dentro
  do shared kernel — o critério usado aqui foi: só entra no shared
  kernel o que é idêntico por construção em todo bounded context, nunca
  o que apenas parece igual hoje.
- Consumidores do `output.OutboxRepository` de cada serviço (mocks de
  teste, por exemplo) agora precisam implementar os três métodos do
  contrato compartilhado (`SaveEvents`, `FetchUnprocessed`,
  `MarkProcessed`), mesmo que ainda não usem os dois últimos.

## Consequences

    shared/
      events/       (BaseEvent, DomainEvent)
      messaging/kafka/  (Producer)
      outbox/       (Repository, PostgresRepository)
      postgres/     (DB, NewConnection)
        ↑                ↑                ↑
    order-service    inventory-service    cdc-connector
    (domain/event/*  (internal/domain/    (relay.OutboxReader,
     compõe           event/* compõe      relay.Publisher — usa
     shared.BaseEvent) shared.BaseEvent)  outbox e kafka direto)

Todo novo bounded context (analytics-service, o futuro user/admin)
deve compor `shared/events.BaseEvent` em seus eventos de domínio e, se
adotar o Transactional Outbox, usar `shared/outbox.Repository` através
de sua própria porta `application/port/output.OutboxRepository`.

## Update (2026-09-19) — shared/postgres e remoção do EventPublisher

Ao implementar o `cdc-connector` (ADR-006), um quarto pacote entrou no
shared kernel pela regra dos três: `shared/postgres`, com o `DB`/
`NewConnection` que order-service e inventory-service já tinham cada
um a sua cópia idêntica, e que o cdc-connector também precisava. Cada
serviço mantém seu próprio `postgres.DB`/`NewConnection` locais como
alias de tipo para `shared/postgres`, seguindo o mesmo padrão já usado
para `output.OutboxRepository`.

No mesmo momento, `EventPublisher` e `Serializer` foram removidos de
`shared/messaging/kafka`. Eles nunca chegaram a ser usados: a interface
`Event` (`EventType()`/`ResourceID()`) que `EventPublisher.Publish`
esperava não é implementada por nenhum evento de domínio real (que
implementam `events.DomainEvent`, com `EventName()`/`AggregateId()`) —
uma discrepância que só ficou visível quando o `cdc-connector` foi
escrito e revelou que a publicação real não parte de um evento Go a
serializar, e sim de uma linha do outbox cujo `Payload` já é o JSON
serializado desde que a aplicação gravou o evento (ver ADR-003,
LEARNING-004). O único primitivo realmente necessário é
`Producer.Publish(topic, key, value []byte)`, que continua em
`shared/messaging/kafka`.

## Security Considerations

Nenhuma mudança de superfície de segurança: o shared kernel não expõe
endpoints nem lida diretamente com credenciais. As considerações de
segurança de ADR-003 (payload de eventos em texto claro) e ADR-004
(TLS/autenticação no Kafka) continuam valendo, agora centralizadas em
um único ponto de implementação.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-004 — Uso de Kafka
- ADR-006 — cdc-connector como relay de polling do Outbox
