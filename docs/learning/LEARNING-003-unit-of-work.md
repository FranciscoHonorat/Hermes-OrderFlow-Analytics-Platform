# LEARNING-003 — Unit of Work

- Date: 2026-09-18
- Topic: Unit of Work
- Related ADR: ADR-002, ADR-003

## Initial Hypothesis

A hipótese inicial era que Unit of Work fosse apenas "uma transação do
banco encapsulada em uma interface", útil principalmente para não
espalhar `BEGIN`/`COMMIT` pelo código de infraestrutura.

## What I Expected

Esperava que a Unit of Work servisse para:

- abrir e fechar uma transação por caso de uso;
- garantir rollback em caso de erro;
- pouco mais que isso.

## What I Learned

Implementar a Unit of Work do order-service (`postgres.unitOfWork`) e
depois replicá-la para o inventory-service deixou claro que o papel
real dela é outro: ela é o que torna o Transactional Outbox (ADR-003)
possível.

A Unit of Work não apenas abre uma transação — ela constrói um
`RepositoryProvider` que entrega, dentro dessa mesma transação, tanto o
repositório do agregado quanto o repositório do outbox:

    UnitOfWork.Do(ctx, func(store RepositoryProvider) error {
        store.StockRepository().Save(ctx, item)
        store.OutboxRepository().SaveEvents(ctx, item.PullEvents())
    })

Isso só funciona porque `repositoryProvider` guarda a `pgx.Tx` e
constrói os dois repositórios (`StockRepositoryFromTx`,
`outbox.NewPostgresRepository(tx)`) a partir dela. Se cada repositório
abrisse sua própria conexão, a atomicidade entre salvar o agregado e
gravar o evento — o ponto inteiro do ADR-003 — deixaria de existir.

## Evidence

Ao portar a Unit of Work para o inventory-service
(`internal/infrastructure/persistence/postgres/unit_of_work.go`), o
padrão se repetiu de forma quase mecânica:

    func (u *unitOfWork) Do(ctx, fn) error {
        tx := db.Pool.Begin(ctx)
        provider := &repositoryProvider{tx: tx, mapper: NewStockMapper()}
        err := fn(provider)
        if err != nil { tx.Rollback(ctx); return err }
        return tx.Commit(ctx)
    }

O fato de essa estrutura ser idêntica entre order-service e
inventory-service — só troca o repositório específico do agregado — foi
o que motivou mover o repositório do outbox para `shared/outbox`
(ADR-005): a Unit of Work em si continua local a cada serviço (porque o
`RepositoryProvider` expõe um repositório de agregado diferente por
bounded context), mas a peça do outbox dentro dela não precisava ser
reescrita a cada novo serviço.

## What Changed In My Understanding

Minha compreensão mudou de:

"Unit of Work é uma transação encapsulada"

para:

"Unit of Work é o mecanismo que garante que o agregado e seus eventos
de domínio cheguem ao banco atomicamente — a transação é apenas o meio,
não o fim."

## Trade-offs I Learned

- A Unit of Work acopla a aplicação a uma implementação transacional
  específica (aqui, `pgx.Tx`) através da interface `RepositoryProvider`
  — um preço aceitável, já que a atomicidade exigida pelo outbox não
  existe sem uma transação real por trás.
- Cada novo agregado exige um novo `RepositoryProvider`/`UnitOfWork`;
  não há como generalizar totalmente esse ponto sem generics pesados,
  então a duplicação estrutural (não de comportamento) entre serviços é
  esperada e aceitável.

## Open Questions

- Como a Unit of Work deve se comportar quando o caso de uso precisa
  interagir com mais de um agregado (ex: uma futura saga dentro do
  mesmo serviço)?
- Vale a pena extrair um helper genérico para o boilerplate de
  `Begin`/`Rollback`/`Commit`, já que ele é idêntico entre
  order-service e inventory-service?

## Next Experiment

Ao implementar o `cdc-connector` (Fase 2 do roadmap), investigar se ele
deve interagir com a tabela outbox de cada serviço via uma conexão
direta de leitura, ou se cada serviço deveria expor isso de outra
forma — essa decisão tem implicação direta em quão "compartilhada" a
Unit of Work de cada bounded context pode continuar sendo.
