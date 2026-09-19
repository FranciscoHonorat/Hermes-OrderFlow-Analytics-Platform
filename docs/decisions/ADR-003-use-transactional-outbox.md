# ADR-003 — Uso do padrão Transactional Outbox

- Status: Accepted
- Date: 2026-09-18

## Context

O Hermes OrderFlow é uma plataforma orientada a eventos: mudanças de
estado relevantes em um bounded context (ex: um pedido colocado, um
estoque reservado) precisam ser publicadas como eventos de domínio para
que outros serviços reajam a elas via Kafka.

O problema clássico desse cenário é a inconsistência entre dois sistemas
que não compartilham a mesma transação:

    salvar Order no PostgreSQL
        ↓
    publicar OrderPlaced no Kafka

Se a aplicação falhar entre os dois passos — ou se o broker estiver
indisponível no momento da publicação — o estado do banco e o estado dos
consumidores de eventos divergem silenciosamente. Não é possível
garantir atomicidade entre um COMMIT no PostgreSQL e um ACK do Kafka.

## Decision

Foi adotado o padrão Transactional Outbox: em vez de publicar o evento
diretamente no Kafka dentro do caso de uso, o evento é gravado em uma
tabela `outbox` na mesma transação que persiste o agregado.

    BEGIN
      ↓
    salvar Order
      ↓
    salvar OrderPlaced na tabela outbox
      ↓
    COMMIT

Um processo separado — o `cdc-connector` — é responsável por ler as
linhas não processadas da tabela `outbox` e publicá-las no Kafka,
marcando-as como processadas após a confirmação do broker. Esse processo
ainda não está implementado; a persistência do outbox foi construída
para sustentá-lo.

O contrato do outbox (`Row`, `Repository`) e sua implementação em
PostgreSQL vivem em `shared/outbox`, reaproveitados por todo bounded
context que precise publicar eventos de domínio — ver ADR-005.

## Why Transactional Outbox

A alternativa mais simples — publicar o evento no Kafka logo após o
commit do banco — introduz uma janela em que a aplicação pode falhar
depois de persistir a Order e antes de publicar o evento,
resultando em um evento de domínio permanentemente perdido.

Gravar o evento na mesma transação do agregado elimina essa janela: ou
os dois são persistidos juntos, ou nenhum é. A publicação real no Kafka
passa a ser um problema de "at-least-once delivery" resolvido de forma
assíncrona pelo relay, e não mais um problema de atomicidade
transacional.

## Alternatives Considered

### Publicar diretamente no Kafka dentro do caso de uso

Mais simples de implementar, mas reintroduz o problema de
inconsistência descrito no contexto: não há forma de garantir que a
gravação no banco e a publicação no broker aconteçam atomicamente.

### Change Data Capture via log de replicação (Debezium)

Ler o WAL do PostgreSQL diretamente (log-based CDC) evita a necessidade
de uma tabela outbox dedicada e captura qualquer mudança na tabela de
origem. Foi descartado para a primeira versão por adicionar
complexidade operacional (Kafka Connect, conectores Debezium) que não
se justifica no estágio atual do projeto. A tabela outbox com um relay
por polling é o ponto de partida; migrar para CDC baseado em WAL fica
registrado como evolução possível do `cdc-connector`.

## Trade-offs

### Benefícios

- Atomicidade entre a mudança de estado do agregado e o registro do
  evento, usando apenas as garantias transacionais que o PostgreSQL já
  oferece (ver ADR-001).
- Desacopla a disponibilidade do Kafka da disponibilidade do caso de
  uso: mesmo com o broker fora do ar, o evento fica persistido e será
  entregue quando o relay conseguir publicá-lo.
- O contrato é simples o suficiente para ser reaproveitado por qualquer
  bounded context (order-service, inventory-service, e os que vierem a
  seguir).

### Custos

- Introduz um processo adicional (`cdc-connector`) responsável por
  drenar a tabela outbox — enquanto ele não existir, os eventos
  gravados na tabela nunca são publicados.
- Entrega é at-least-once: o relay pode publicar o mesmo evento mais de
  uma vez em caso de falha entre a publicação e o `MarkProcessed`.
  Consumidores precisam ser idempotentes.
- A tabela outbox cresce indefinidamente sem uma rotina de limpeza das
  linhas já processadas.

## Consequences

Cada bounded context que produz eventos de domínio passa a ter sua
própria tabela `outbox`, criada via migration
(`migrations/000002_create_outbox.up.sql`), e grava nela usando
`shared/outbox.Repository` dentro da mesma unit of work que persiste o
agregado.

    Application
        ↓
    UnitOfWork.Do
        ↓
    StockRepository().Save(item)
    OutboxRepository().SaveEvents(item.PullEvents())
        ↓
    COMMIT

O `cdc-connector` (ainda não implementado) usará `FetchUnprocessed` e
`MarkProcessed`, também definidos em `shared/outbox.Repository`, para
drenar essa tabela e publicar no Kafka.

## Security Considerations

O payload do evento é serializado como JSON e armazenado em texto claro
na coluna `payload`. Eventos de domínio não devem carregar dados
sensíveis (senhas, tokens, dados de pagamento) sem antes avaliar
necessidade de criptografia em repouso ou mascaramento de campos.

## Related Decisions

- ADR-001 — Uso do PostgreSQL
- ADR-002 — Uso de Hexagonal Architecture
- ADR-004 — Uso de Kafka
- ADR-005 — Shared Kernel entre bounded contexts
