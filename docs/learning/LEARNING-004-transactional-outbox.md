# LEARNING-004 — Transactional Outbox

- Date: 2026-09-18
- Topic: Transactional Outbox
- Related ADR: ADR-003, ADR-004, ADR-005

## Initial Hypothesis

A hipótese inicial era que implementar o Transactional Outbox
significava, basicamente, trocar "publicar no Kafka" por "inserir numa
tabela outbox" — um detalhe de implementação isolado dentro do
order-service.

## What I Expected

Esperava que, ao terminar de implementar `OutboxRepository.SaveEvents`
no order-service, o padrão estivesse "pronto": bastaria replicar a
mesma tabela e o mesmo código em qualquer serviço novo.

## What I Learned

Ao começar o inventory-service, ficou claro que a metade que eu havia
implementado (`SaveEvents`, dentro da transação) é só a escrita do
padrão. A outra metade — ler as linhas não processadas e publicá-las no
Kafka — nunca tinha sido escrita: o `cmd/server/main.go` do
order-service já instanciava um `outboxRepo`, mas ele ficava sem uso
(`_ = outboxRepo`), porque não havia nenhum processo lendo a tabela.

Isso reformulou como entendo o padrão:

    Escrita (dentro da transação da aplicação)
        SaveEvents(ctx, evts)
            ↓
    outbox (tabela)
            ↓
    Leitura (processo externo, assíncrono)
        FetchUnprocessed(ctx, limit)
            ↓
        publicar no Kafka
            ↓
        MarkProcessed(ctx, ids)

O padrão não está completo enquanto só o lado da escrita existir. A
tabela outbox sem um relay é só uma fila que nunca é drenada.

## Evidence

O contrato `shared/outbox.Repository` foi desenhado com essa divisão
explícita:

    type Repository interface {
        SaveEvents(ctx, evts []events.DomainEvent) error
        FetchUnprocessed(ctx, limit int) ([]Row, error)
        MarkProcessed(ctx, ids []uuid.UUID) error
    }

`SaveEvents` é chamado pela aplicação, dentro da Unit of Work (ver
LEARNING-003). `FetchUnprocessed` e `MarkProcessed` existem para o
`cdc-connector`, que ainda não foi implementado — a tabela outbox do
order-service e do inventory-service já está pronta para ele, mas
nenhum evento gravado hoje chega de fato ao Kafka.

Isso também expôs por que a mensageria não podia ficar duplicada por
serviço (ver ADR-005): o relay vai precisar tratar a tabela outbox de
qualquer bounded context da mesma forma, então o contrato de leitura
precisa ser idêntico em todos eles.

## What Changed In My Understanding

Minha compreensão mudou de:

"Transactional Outbox é uma tabela em vez de uma chamada de rede"

para:

"Transactional Outbox é um contrato de duas pontas — escrita atômica e
leitura assíncrona — e o padrão só entrega a garantia que promete
(nenhum evento perdido) quando as duas pontas existem."

## Trade-offs I Learned

- Entrega passa a ser at-least-once: se o relay publicar no Kafka mas
  falhar antes de chamar `MarkProcessed`, o mesmo evento será publicado
  de novo na próxima leitura. Todo consumidor Kafka do sistema precisa
  ser idempotente por causa dessa escolha, não apesar dela.
- A tabela outbox cresce indefinidamente sem uma rotina de purga das
  linhas já processadas — ainda não implementada.
- Até o `cdc-connector` existir, o padrão está "pela metade" em
  produção: os eventos são persistidos de forma segura, mas não
  chegam a lugar nenhum.

## Open Questions

- Qual o intervalo de polling adequado para o relay? Um polling
  agressivo demais gera carga desnecessária no Postgres; um polling
  esparso demais aumenta a latência de entrega dos eventos.
- Em que ponto compensa migrar de polling para CDC baseado em log
  (Debezium/WAL), como já cogitado na ADR-003?

## Next Experiment

Implementar o `cdc-connector` (Fase 2 do roadmap) como um processo de
polling simples sobre `FetchUnprocessed`/`MarkProcessed`, primeiro
contra a tabela outbox do order-service, depois generalizado para
qualquer bounded context — e medir, na prática, a latência entre o
`SaveEvents` e a publicação efetiva no Kafka.
