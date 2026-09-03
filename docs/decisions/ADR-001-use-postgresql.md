# ADR-001 — Uso do PostgreSQL

- Status: Accepted
- Date: 2026-08-30

## Context

O Hermes OrderFlow é uma plataforma educacional de logística e gestão de
pedidos, com um domínio financeiro.

O Order Service precisa persistir informações relacionadas aos pedidos,
incluindo:

- identificador do pedido;
- cliente;
- valor total;
- status;
- timestamps;
- eventos de domínio.

Durante o desenvolvimento também foi adotado o padrão Transactional
Outbox, no qual a alteração da Order e o registro de um evento precisam
participar da mesma transação.

Inicialmente, a escolha do banco foi orientada principalmente pela
adequação percebida do PostgreSQL para aplicações de e-commerce e pela
necessidade de uma base relacional para os dados do pedido.

## Decision

Foi escolhido o PostgreSQL como banco de dados do Order Service.

A aplicação utilizará PostgreSQL diretamente através de `pgx`, sem
utilizar um ORM.

As alterações estruturais do banco serão controladas por migrations
versionadas utilizando `golang-migrate`.

## Why PostgreSQL

A escolha está relacionada principalmente às características
transacionais necessárias para o domínio.

O Order Service precisa executar operações como:

    BEGIN
      ↓
    salvar Order
      ↓
    salvar Outbox Event
      ↓
    COMMIT

Caso uma das operações falhe:

    ROLLBACK

O PostgreSQL fornece o modelo transacional necessário para manter essas
operações dentro de uma mesma unidade de trabalho.

Além disso, o modelo relacional se encaixa naturalmente nas entidades
persistidas atualmente pelo Order Service.

## Alternatives Considered

### MongoDB

MongoDB poderia ser utilizado para persistência de documentos e oferece
um modelo mais flexível de schema.

Entretanto, para o estágio atual do projeto, foi considerado mais
interessante utilizar um banco relacional para estudar explicitamente
transações, constraints, índices e consistência.

### MySQL

MySQL também atenderia aos requisitos básicos de persistência
relacional.

Entretanto, PostgreSQL foi escolhido como a tecnologia principal do
projeto, considerando sua adequação ao modelo transacional e ao
objetivo educacional de aprofundar o conhecimento sobre bancos
relacionais.

## Trade-offs

### Benefícios

- Modelo relacional adequado para os dados atuais.
- Suporte a transações.
- Constraints e integridade estrutural.
- Índices.
- Tipos como UUID, JSONB e TIMESTAMPTZ.
- Bom suporte através do `pgx`.
- Permite trabalhar diretamente com SQL.
- Adequado para estudar persistência transacional.

### Custos

- Exige modelagem explícita do schema.
- Alterações estruturais precisam ser planejadas e versionadas.
- Escalabilidade horizontal exige decisões arquiteturais adicionais.
- O uso de SQL explícito aumenta a responsabilidade do desenvolvedor
  sobre queries, índices e performance.

## Consequences

O PostgreSQL passa a ser a fonte de persistência do Order Service.

O schema será versionado no repositório:

    services/order-service/migrations/

As operações de persistência utilizarão SQL explícito através do `pgx`.

A aplicação poderá utilizar transações para garantir atomicidade entre
operações relacionadas, especialmente entre a persistência da Order e
seu registro na Outbox.

A decisão também aumenta a responsabilidade da equipe sobre:

- modelagem;
- migrations;
- queries;
- índices;
- transações;
- concorrência;
- performance.

## Security Considerations

A conexão com o PostgreSQL deve utilizar credenciais armazenadas fora
do código-fonte.

Em ambientes de produção, a comunicação com o banco deve ser protegida
e o usuário da aplicação deve possuir somente as permissões necessárias.

Credenciais padrão utilizadas no ambiente local do Docker Compose são
exclusivas para desenvolvimento.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-004 — Uso de Kafka