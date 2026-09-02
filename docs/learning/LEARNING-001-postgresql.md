# LEARNING-001 — PostgreSQL

- Date: 2026-08-30
- Topic: PostgreSQL
- Related ADR: ADR-001

## Initial Hypothesis

Inicialmente, PostgreSQL foi escolhido porque parecia ser uma boa opção
para uma plataforma de e-commerce.

A hipótese inicial era baseada principalmente na adequação percebida de
um banco relacional para o domínio de pedidos.

## What I Expected

Esperava utilizar o PostgreSQL principalmente para:

- persistir Orders;
- realizar consultas;
- manter os dados estruturados;
- utilizar transações quando necessário.

## What I Learned

Durante a implementação do Order Service, percebi que a escolha do banco
estava relacionada a uma questão mais importante do que apenas
persistência.

A implementação do Transactional Outbox criou a necessidade de persistir
a Order e o evento relacionado dentro da mesma transação.

Isso tornou conceitos como:

- BEGIN;
- COMMIT;
- ROLLBACK;
- atomicidade;
- constraints;
- SQL explícito;

relevantes para a decisão arquitetural.

## Evidence

Foi implementado um fluxo em que:

    BEGIN
      ↓
    Save Order
      ↓
    Save Outbox Event
      ↓
    COMMIT

Caso uma das operações falhe:

    ROLLBACK

O comportamento foi validado utilizando PostgreSQL, pgx e Unit of Work.

## What Changed In My Understanding

Minha compreensão mudou de:

"PostgreSQL é adequado para e-commerce"

para:

"PostgreSQL atende ao modelo relacional e transacional necessário para
o contexto atual, especialmente quando precisamos garantir atomicidade
entre operações relacionadas."

## Trade-offs I Learned

O uso de PostgreSQL não elimina decisões arquiteturais.

Ainda precisamos considerar:

- modelagem;
- índices;
- queries;
- concorrência;
- locks;
- isolamento;
- migrations;
- performance;
- escalabilidade.

## Open Questions

- Como o PostgreSQL se comportará sob alta concorrência?
- Como escolher índices com base em evidências?
- Como funcionam MVCC e locks internamente?
- Como diferentes níveis de isolamento afetam o Order Service?
- Quando particionamento seria necessário?

## Next Experiment

Investigar o comportamento de transações e concorrência no PostgreSQL
antes de implementar cenários mais complexos no Financial Service.