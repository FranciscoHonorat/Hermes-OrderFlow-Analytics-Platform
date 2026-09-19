# Contributing

Obrigado por contribuir com o Hermes OrderFlow Analytics Platform.

Este projeto é um projeto educacional voltado ao estudo e à aplicação prática de:

- Go
- Domain-Driven Design
- Hexagonal Architecture
- Sistemas Distribuídos
- PostgreSQL
- Kafka
- Segurança
- Sistemas financeiros

O objetivo não é apenas produzir código funcional, mas desenvolver decisões
técnicas justificáveis, compreender trade-offs e construir um sistema
financeiro distribuído de forma incremental.

---

## 1. Filosofia do Projeto

O projeto segue alguns princípios fundamentais:

- Informação não é conhecimento.
- Toda decisão técnica deve possuir uma justificativa.
- Hipóteses devem ser formuladas antes da implementação.
- Trade-offs devem ser explicitados.
- O domínio não deve depender de detalhes de infraestrutura.
- Falhas parciais devem ser consideradas.
- Código funcional não significa código arquiteturalmente correto.
- Testes devem validar comportamento e invariantes.
- Segurança deve ser considerada desde o desenho.
- Mudanças arquiteturais devem ser documentadas.

O objetivo final é desenvolver a capacidade de tomar melhores decisões
técnicas.

---

## 2. Estrutura do Projeto

A estrutura principal segue uma separação por serviços:

services/
├── order-service/
├── payment-service/
├── ledger-service/
├── risk-service/
├── reconciliation-service/
└── ...

Cada serviço deve possuir suas próprias responsabilidades e fronteiras.

No `order-service`, a estrutura segue aproximadamente:

application/
├── command/
├── query/
└── port/

domain/
├── entity/
├── event/
├── repository/
├── valueobject/
└── ...

infrastructure/
├── postgres/
├── kafka/
└── ...

migrations/

---

## 3. Arquitetura

O projeto utiliza conceitos de:

- Domain-Driven Design
- Hexagonal Architecture
- Ports & Adapters
- Repository Pattern
- Unit of Work
- Domain Events
- Transactional Outbox

A dependência deve apontar para dentro da aplicação e do domínio.

O domínio não deve conhecer:

- PostgreSQL
- Kafka
- HTTP
- Docker
- Redis
- bibliotecas específicas de infraestrutura

Esses detalhes devem permanecer nos adapters.

---

## 4. Regras de Arquitetura

### Domain

O domínio contém regras de negócio e invariantes.

Não deve depender de infraestrutura.

### Application

A camada de aplicação coordena os casos de uso.

Ela pode:

- receber comandos;
- coordenar repositories;
- controlar Unit of Work;
- disparar operações de domínio;
- coordenar persistência de eventos.

Ela não deve concentrar regras que pertencem ao domínio.

### Infrastructure

A infraestrutura implementa os ports definidos pela aplicação.

Exemplos:

- PostgreSQL
- Kafka
- HTTP
- Redis

---

## 5. Invariantes

Toda alteração relevante deve considerar as invariantes do projeto.

### Conservação

O estado financeiro não pode ser criado ou destruído
indevidamente.

### Idempotência

Reprocessamentos não devem produzir efeitos indevidos.

### Finalidade

Uma operação finalizada deve possuir um estado claramente definido.

### Não Repúdio

Operações relevantes devem possuir evidências suficientes
para demonstrar o que ocorreu.

### Confidencialidade

Informações sensíveis devem ser protegidas.

### Autorização

Operações devem ser executadas somente por atores autorizados.

Nenhuma mudança arquitetural deve quebrar essas invariantes sem
uma justificativa explícita e documentada.

---

## 6. Banco de Dados

O banco de dados é PostgreSQL.

O schema oficial deve ser definido através de migrations versionadas.

As migrations ficam em:

services/<service>/migrations/

Exemplo:

000001_create_orders.up.sql
000001_create_orders.down.sql

Não utilizar o DBeaver como fonte oficial do schema.

Alterações manuais realizadas diretamente no banco não substituem
uma migration versionada.

---

## 7. Migrations

Toda alteração estrutural no banco deve possuir migration.

Exemplo:

000002_create_outbox.up.sql
000002_create_outbox.down.sql

A migration `up` deve aplicar a alteração.

A migration `down` deve permitir a reversão quando isso for
tecnicamente seguro.

Antes de criar uma migration deve-se verificar:

- dependências;
- constraints;
- índices;
- compatibilidade;
- impacto sobre dados existentes;
- rollback;
- concorrência.

---

## 8. ORM

O projeto utiliza SQL explícito através de `pgx`.

Não utilizar ORM para esconder operações importantes do banco.

A intenção é manter explícito o conhecimento sobre:

- SQL;
- transações;
- índices;
- constraints;
- locks;
- isolamento;
- performance;
- comportamento do PostgreSQL.

---

## 9. Transactions

Operações que precisam ser atomicamente consistentes devem utilizar
uma transação.

O projeto utiliza Unit of Work para controlar esse processo.

Exemplo conceitual:

BEGIN

    salvar aggregate
    salvar eventos na outbox

COMMIT

Caso alguma operação falhe:

ROLLBACK

A persistência do estado e dos eventos que pertencem à mesma operação
de negócio deve ocorrer dentro da mesma transação quando a consistência
exigir atomicidade.

---

## 10. Transactional Outbox

Eventos que precisam sobreviver a falhas entre o banco e o broker
devem utilizar o padrão Outbox.

Fluxo:

Order
  ↓
Domain Event
  ↓
Outbox
  ↓
COMMIT
  ↓
Outbox Relay
  ↓
Kafka

Não publicar diretamente no Kafka antes de garantir a persistência
transacional necessária.

O processamento posterior dos eventos deve considerar:

- retry;
- duplicação;
- idempotência;
- falhas parciais.

---

## 11. Código

O código Go deve seguir as convenções idiomáticas da linguagem.

Antes de enviar uma alteração:

```bash
go fmt ./...
go test ./...
go mod tidy
