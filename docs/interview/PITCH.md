# Pitch técnico — OrderFlow Analytics Platform

Material de apoio para apresentar este projeto em entrevista técnica.

---

## Elevator pitch (30-60s)

> "O OrderFlow é uma plataforma de gestão de pedidos com pipeline analítico
> em tempo real, que eu construí pra praticar arquitetura de sistemas
> distribuídos de propósito: separação OLTP/OLAP, event-driven architecture
> com transactional outbox, e serviços em Go organizados com Clean
> Architecture e DDD. Não é um produto pronto pra produção ainda,
> mas cada decisão de arquitetura documentadas em ADRs e
> pelo menos um bug real de infraestrutura que eu diagnostiquei e corrigi
> validando contra o stack real, não só contra testes automatizados."


---

## Arquitetura em uma frase por serviço

- **order-service** — contexto principal de pedidos; grava eventos de
  domínio na própria tabela outbox na mesma transação do agregado.
- **inventory-service** — gestão de estoque, mesma arquitetura hexagonal e
  outbox do order-service (endpoint de criação de SKU foi deliberadamente
  adiado — decisão de escopo, não lacuna esquecida).
- **cdc-connector** — processo único que drena a tabela outbox de múltiplos
  bounded contexts por polling e publica no Kafka.
- **user-service** — autenticação e autorização via JWT, propagado para
  order/inventory/frontend.
- **shared/** — kernel com os contratos de outbox e eventos de domínio,
  reaproveitado por todo bounded context.

---

## Três histórias técnicas (formato: contexto → decisão → alternativa descartada → custo → resultado)

### 1. Kafka em crash-loop silencioso

**Contexto concreto** — Depois de implementar o cdc-connector, o relay
"rodava" normalmente (healthcheck verde, sem log de erro), mas nenhum
evento chegava ao Kafka. O container do broker tinha `RestartCount: 31` —
estava em crash-loop desde sempre, e ninguém tinha percebido porque nenhum
serviço antes do cdc-connector se conectava de fato ao broker.

**Decisão tomada** — Corrigir `KAFKA_ADVERTISED_LISTENERS`: separar
`PLAINTEXT` (porta 9092, tráfego interno do compose) de `PLAINTEXT_HOST`
(porta 9093, acesso do host) — as duas apontavam pra mesma porta e o broker
recusava subir ("two listeners on the same port").

**Alternativa descartada** — Confiar só nos testes de integração com
testcontainers. Testcontainers sobe seu próprio Kafka isolado com config
própria — nunca exerceria esse bug específico do `docker-compose.yml`.

**Custo pago** — Diagnóstico caro: o `ProduceSync` do cliente `kgo` ficava
bloqueado indefinidamente esperando o broker, só retornando com `context
canceled` no shutdown. "Relay travado" parecia, de fora, idêntico a "relay
sem nada pra processar".

**Resultado observado** — Depois do fix, validado contra o stack real do
compose, os eventos passaram a fluir order-service → outbox → Kafka de
ponta a ponta. Virou prática registrada: sempre validar mensageria contra o
compose real antes de considerar a feature pronta.

### 2. Transactional Outbox (ADR-003)

**Contexto concreto** — O fluxo ingênuo seria `salvar Order no Postgres` →
`publicar OrderPlaced no Kafka`. Se a aplicação cai entre os dois passos, ou
o broker está indisponível, o evento se perde silenciosamente — não existe
transação que cubra banco e broker ao mesmo tempo.

**Decisão tomada** — Gravar o evento numa tabela `outbox` na mesma
transação que persiste o agregado; um processo separado (cdc-connector)
drena essa tabela e publica no Kafka depois.

**Alternativa descartada** — CDC baseado em log de replicação (Debezium
lendo o WAL do Postgres). Descartado na v1 por trazer complexidade
operacional (Kafka Connect, replication slots) desproporcional ao estágio
do projeto.

**Custo pago** — Entrega at-least-once (consumidores precisam ser
idempotentes), tabela outbox cresce sem rotina de limpeza ainda
implementada, e a indisponibilidade do cdc-connector pausa a entrega (sem
perder dado, mas atrasando).

**Resultado observado** — Atomicidade entre estado do agregado e registro
do evento usando só garantias que o Postgres já dá. Contrato reaproveitado
sem duplicação entre order-service e inventory-service via shared kernel.

### 3. Relay por polling em vez de CDC via log

**Contexto concreto** — Com o outbox escrevendo desde a ADR-003, nada ainda
lia a tabela — os eventos ficavam presos, `FetchUnprocessed`/`MarkProcessed`
existiam como contrato sem chamador real.

**Decisão tomada** — cdc-connector v1 como relay de polling: uma goroutine
roda a cada `POLL_INTERVAL` (padrão 2s), busca não-processados por source
configurada via env `SOURCES`, publica e marca como processado.

**Alternativa descartada** — Um relay por bounded context em vez de
processo central: evitaria compartilhar credenciais entre bancos, mas
duplicaria o loop de polling e a lógica de retry/backoff N vezes.

**Custo pago** — Latência de entrega limitada pelo intervalo de poll (2s,
não near-real-time); processo central com acesso a múltiplos bancos é
ponto único de acoplamento operacional; custo de poll no Postgres mesmo sem
eventos novos.

**Resultado observado** — Fechou o ciclo do outbox de ponta a ponta;
adicionar um novo bounded context como source virou uma linha na variável
`SOURCES`, sem mudar código; lógica de negócio do relay testável isolada de
I/O real via interfaces pequenas (`OutboxReader`, `Publisher`).

---

## O que reconhecer como limitação (sem esconder)

- Projeto em desenvolvimento, não é produção real.
- Endpoint de "criar SKU" no inventory-service foi deliberadamente adiado.
- Observabilidade (Prometheus/Grafana) e Kubernetes estão no roadmap, não
  totalmente implementados — ser honesto sobre o que já existe vs. o que
  está planejado.
- Tabela outbox ainda não tem rotina de limpeza das linhas processadas.

## Perguntas prováveis do entrevistador

- Por que Go? Por que Kafka e não outro broker?
- Como você garante idempotência no consumo de eventos?
- Como testar um sistema orientado a eventos? (relay testado com fakes +
  testcontainers para o caminho real)
- Por que não Debezium/CDC baseado em log desde o início?
- O que você faria diferente se recomeçasse agora?
