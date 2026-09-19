# Guia Técnico Aprofundado — inventory-service e cdc-connector

Este documento detalha, tópico a tópico, cada tecnologia, biblioteca e
técnica usada na construção do `inventory-service` e do `cdc-connector`,
com a justificativa de escolha e os trade-offs envolvidos.

Ele é complementar a três outras fontes que não repete:

- [`ARCHITECTURE.md`](../../ARCHITECTURE.md) — o conjunto de regras
  *prescritivas* que todo o projeto deve seguir (o "como deve ser").
- `docs/decisions/ADR-*.md` — as decisões pontuais e suas alternativas
  descartadas (o "por que decidimos X e não Y", registrado no momento da
  decisão).
- `docs/learning/LEARNING-*.md` — o que a implementação real ensinou,
  incluindo hipóteses erradas e bugs encontrados (o "o que a prática
  mostrou que a decisão não previu").

Aqui o objetivo é diferente: uma explicação de referência, aprofundada,
de cada peça da stack e de cada técnica — como se fosse a "planta baixa"
completa para alguém entrando no projeto agora.

---

## Parte 1 — Stack tecnológica

### 1.1 Go + workspace multi-módulo (`go.work`)

**O que é.** Go workspaces (`go.work`) permitem que múltiplos módulos Go
(`go.mod` separados) sejam desenvolvidos e testados juntos localmente,
sem precisar publicar/versionar cada um. Cada serviço (`shared`,
`order-service`, `inventory-service`, `cdc-connector`) é um módulo
próprio; o `go.work` na raiz lista quais módulos participam do
workspace via `use (...)`.

**Por que aqui.** O projeto é um monorepo com múltiplos bounded
contexts (ADR-002) que precisam compartilhar código de infraestrutura
comum (`shared/outbox`, `shared/events`, `shared/messaging/kafka`,
`shared/postgres` — ADR-005) sem acoplar os *domínios* entre si. Cada
serviço declara `require github.com/.../shared ...` normalmente, mas
usa `replace github.com/.../shared => ../../shared` para resolver
contra o caminho local em vez de buscar uma versão publicada.

**Trade-offs.**
- ✅ Cada serviço mantém seu próprio `go.mod` — dependências
  divergentes (ex: `gin` só no inventory-service) não poluem os outros
  serviços, e cada um pode em teoria evoluir sua versão de Go
  independentemente (na prática hoje 1.25.11 vs 1.26.0, uma divergência
  pequena e não intencional).
- ✅ Sem necessidade de publicar `shared` num registry privado nem de
  gerenciar tags de versão para código interno que muda o tempo todo.
- ❌ `go <cmd> ./...` não atravessa módulos a partir da raiz do repo —
  por isso o `Makefile` precisa de um loop explícito sobre `MODULES`
  para `build`/`test`/`fmt`/`tidy` (comentário no próprio Makefile
  documenta essa armadilha).
- ❌ O `replace` local é uma dependência implícita da estrutura de
  diretórios (`../../shared`) — mover um serviço de lugar quebra o
  build até ajustar o `go.mod`.
- ⚠️ Versões de Go divergentes entre módulos (1.25.11 vs 1.26.0) não
  quebram nada hoje, mas é um sinal de que o `go.mod tidy` não está
  sendo rodado de forma uniforme entre serviços — vale alinhar.

### 1.2 Gin (`github.com/gin-gonic/gin`)

**O que é.** Framework HTTP minimalista para Go, com roteamento, bind
de JSON e middlewares.

**Por que aqui.** Usado só no `inventory-service` (e order-service) para
expor a API REST de comandos (`POST /stock/reserve`, etc.) e queries.
`ShouldBindJSON` faz o parse+validação de payload em uma linha
(`stock_handler.go`), delegando o corpo da lógica para os use cases —
o handler nunca tem regra de negócio, só tradução HTTP ↔ domínio.

**Por que não no cdc-connector.** Ele não tem API de negócio, só um
endpoint `/health` — por isso usa `net/http` puro (`http.ServeMux`) em
vez de trazer Gin como dependência para uma única rota trivial.

**Trade-offs.**
- ✅ Gin é rápido, maduro, com bind/validation embutidos
  (`go-playground/validator` por baixo dos panos) — menos código
  boilerplate de parsing manual.
- ✅ Router baseado em árvore de radix — overhead de roteamento
  desprezível mesmo com muitas rotas.
- ❌ Acopla a camada HTTP a uma dependência externa específica (embora
  isolada em `infrastructure/http`, seguindo a arquitetura hexagonal —
  trocar por outro framework não afeta domínio/aplicação).
- ❌ Traz uma árvore de dependências transitivas relativamente grande
  (`bytedance/sonic` para JSON rápido via assembly, `go-playground/*`,
  etc.) — justificável para uma API HTTP real, desperdício para um
  processo com uma única rota de health check (daí a escolha de não
  usá-lo no cdc-connector).

### 1.3 PostgreSQL 16 + `jackc/pgx/v5`

**O que é.** PostgreSQL como banco relacional (`postgres:16-alpine`),
com **um schema/database por bounded context** (`orderflow`,
`inventory`) na mesma instância — isolamento lógico, não físico, em
dev (ADR-001). `pgx/v5` é o driver usado — não via `database/sql` (que
abstrairia o driver por baixo de uma interface genérica), mas
diretamente, para ter acesso nativo a `pgxpool.Pool` (pool de conexões)
e `pgx.Tx` (transações explícitas).

**Por que pgx e não `database/sql` + driver.** A escolha direta por
`pgx.Tx` é o que viabiliza o padrão de Unit of Work usado no projeto: o
`unitOfWork.Do` abre uma `pgx.Tx` e a repassa para *múltiplos*
repositórios (`StockRepository` e `OutboxRepository`) dentro do mesmo
escopo transacional, sem nenhuma camada de abstração de
"transação genérica" no meio. `database/sql` também suporta
transações, mas `pgx` nativo evita uma camada extra de tradução de
tipos e expõe funcionalidades específicas do protocolo Postgres
(COPY, tipos nativos) que `database/sql` esconde.

**Trade-offs.**
- ✅ Menos overhead de abstração — SQL cru (`INSERT ... ON CONFLICT DO
  UPDATE`, ver `stock_repository.go`) com mapeamento manual e
  explícito, sem "magia" de ORM.
- ✅ `pgxpool` gerencia pool de conexões nativamente, com boas
  configurações padrão para carga concorrente.
- ❌ Acopla o código de repositório ao `pgx` especificamente — trocar
  de driver Postgres exigiria reescrever os repositórios (mitigado
  pelo fato de que a *interface* de repositório vive no domínio; só a
  implementação em `infrastructure/persistence/postgres` seria
  reescrita).
- ❌ Sem ORM, cada repositório escreve e mantém SQL manualmente — mais
  verboso, mais superfície para erro de digitação em nomes de coluna
  (mitigado por testes de integração reais).
- ⚠️ Múltiplos databases numa única instância Postgres compartilhada é
  uma simplificação de ambiente de dev — em produção, isolar
  fisicamente por serviço (ADR-002 já assume isso como direção).

### 1.4 `golang-migrate` (migrations)

**O que é.** Ferramenta de migração de schema baseada em arquivos SQL
puro versionados (`NNNNNN_nome.up.sql` / `.down.sql`), rodada como
container standalone (`migrate/migrate:v4.17.1`) no `docker-compose.yml`.

**Por que aqui.** Cada serviço tem sua pasta `migrations/` própria
(`services/inventory-service/migrations`), aplicada por um container
efêmero (`inventory-migrate`) antes do serviço subir — orquestrado via
`depends_on: condition: service_completed_successfully`. Migração de
schema é tratada como *infraestrutura declarada*, não como algo que a
aplicação Go faz em runtime na inicialização.

**Trade-offs.**
- ✅ SQL puro e explícito — sem DSL de ORM para gerar migração, o que é
  auditável e revisável em PR como qualquer outro SQL.
- ✅ Separar a migração da aplicação (container próprio) evita que
  múltiplas réplicas do mesmo serviço tentem migrar o schema
  simultaneamente na subida.
- ❌ Mais uma peça móvel no `docker-compose.yml`/pipeline de deploy —
  esquecer de rodar a migração antes do serviço é uma falha de
  orquestração possível (mitigada aqui pelo `depends_on` com condição
  de sucesso).
- ⚠️ Rollback (`down.sql`) escrito manualmente por quem escreve a
  migração — não é gerado automaticamente, então pode divergir do
  `up.sql` se não for mantido com o mesmo cuidado.

### 1.5 Kafka (Confluent + Zookeeper) e `twmb/franz-go`

**O que é.** Apache Kafka como broker de eventos entre bounded
contexts (ADR-004), rodando via imagens Confluent
(`confluentinc/cp-kafka:7.6.1` + `confluentinc/cp-zookeeper:7.6.1`) —
**modo Zookeeper**, não KRaft, no `docker-compose.yml` real. O cliente
usado em Go é `github.com/twmb/franz-go` (pacote `kgo`).

**Por que Zookeeper e não KRaft no compose real.** KRaft (Kafka sem
Zookeeper, usando Raft interno) é o modo mais moderno, e é o que os
testes de integração usam via testcontainers
(`confluentinc/confluent-local`, ADR-006/LEARNING-005). O
`docker-compose.yml` de desenvolvimento, porém, ainda usa o par
Kafka+Zookeeper clássico — uma divergência entre ambiente de teste e
ambiente de dev/prod que foi exatamente o que escondeu o bug de
`KAFKA_ADVERTISED_LISTENERS` (dois listeners na mesma porta) descrito
no LEARNING-005: nenhum teste automatizado exercitava esse
`docker-compose.yml` específico.

**Por que `franz-go` e não `confluent-kafka-go`/`sarama`.**
`franz-go` é um cliente Kafka **puro Go**, sem depender de `cgo` +
`librdkafka` (que é o caso de `confluent-kafka-go`). Isso simplifica
drasticamente o build multi-stage do Dockerfile (`CGO_ENABLED=0`,
cross-compile sem precisar de toolchain C) e o binário final é estático.
`sarama` (Shopify) seria outra opção pura-Go, mas `franz-go` tem uma
API mais moderna e tende a ter melhor desempenho em benchmarks
recentes.

**Detalhe de configuração crítico: `kgo.AllowAutoTopicCreation()`.**
Sem essa opção explícita, o cliente `kgo` **não** pede criação
automática de tópico ao broker — mesmo com
`auto.create.topics.enable=true` no lado do broker. O primeiro
`ProduceSync` num tópico novo falha com `UNKNOWN_TOPIC_OR_PARTITION`.
Esse é um detalhe de protocolo que só aparece testando contra um broker
real (LEARNING-005, bug #1) — testes com um `Publisher` fake nunca o
exercitam.

**Trade-offs.**
- ✅ Cliente puro-Go → build simples, sem dependência de sistema.
- ✅ `ProduceSync` dá uma API simples de "publicar e esperar
  confirmação", adequada ao caso de uso do relay (publicar, só marcar
  como processado após confirmação).
- ❌ `ProduceSync` bloqueia indefinidamente se o broker estiver
  inacessível e não há timeout de produção próprio configurado — só
  responde a cancelamento do `context` do chamador (Open Question
  registrada no LEARNING-005: caso o broker caia, o relay "trava" sem
  logar erro, algo indistinguível de "sem nada para processar" a menos
  que se olhe o processo de perto).
- ⚠️ Zookeeper vs. KRaft divergente entre teste e dev é dívida técnica
  ativa: o `docker-compose.yml` deveria idealmente já estar em KRaft
  (Confluent recomenda isso para novas implantações), alinhado ao que
  os testes de integração já validam.

### 1.6 `google/uuid`

**O que é.** Geração de UUIDs (v4 por padrão) em Go.

**Por que aqui.** Cada linha da tabela `outbox` recebe um `uuid.New()`
como chave primária (`shared/outbox/postgres.go`), desacoplado do ID do
agregado (`AggregateID` é uma coluna separada, usada como chave de
particionamento/deduplicação no Kafka via `key` do `kgo.Record`).

**Trade-offs.**
- ✅ Biblioteca padrão de fato no ecossistema Go, zero dependências
  problemáticas, API simples.
- ⚠️ UUID v4 é aleatório — não é ordenável por tempo de criação (ao
  contrário de um ULID ou UUID v7). Como a tabela outbox já ordena por
  `created_at` explicitamente (`ORDER BY created_at` em
  `FetchUnprocessed`), isso não é um problema prático aqui, mas seria
  uma consideração se o ID precisasse também servir de índice de
  ordenação.

### 1.7 `encoding/json` (stdlib) como serialização de eventos

**O que é.** Os eventos de domínio (`StockReserved`, `LowStockDetected`
etc.) são serializados para `[]byte` via `json.Marshal` puro da
biblioteca padrão, sem lib de serialização externa (Protobuf, Avro,
MessagePack).

**Por que aqui.** O payload gravado na tabela outbox é o mesmo JSON
publicado no Kafka — sem etapa de reserialização no meio. Isso é uma
decisão deliberada registrada no LEARNING-005 (bug #2): havia uma
abstração `EventPublisher`/`Serializer` que reserializaria o evento
antes de publicar, mas isso era desnecessário e a interface nem batia
com o `events.DomainEvent` real — foi removida (aparece como arquivo
deletado no `git status`, `shared/messaging/kafka/event_publisher.go`
e `serializer.go`).

**Trade-offs.**
- ✅ Simplicidade máxima: o que a aplicação grava é literalmente o que
  o consumidor no outro lado do Kafka recebe — sem lugar para
  divergência entre "schema gravado" e "schema publicado".
- ✅ JSON é legível por humanos — depurar eventos direto na tabela
  outbox ou no tópico Kafka (`kcat`/`kafka-console-consumer`) não exige
  ferramenta especial.
- ❌ Sem schema registry nem validação de compatibilidade
  forward/backward — evoluir a forma de um evento (renomear/remover
  campo) pode quebrar consumidores silenciosamente, algo que
  Protobuf+schema registry preveniria em tempo de build/deploy.
- ❌ JSON é mais verboso e mais lento para serializar/desserializar que
  formatos binários — irrelevante no volume atual do projeto, mas seria
  uma reconsideração em alta escala.

### 1.8 `testify` + `testcontainers-go`

**O que é.** `stretchr/testify` para asserções (`assert`/`require`) em
todos os módulos. `testcontainers-go` (+ módulos `postgres` e `kafka`)
para subir dependências reais (Postgres, Kafka em modo KRaft via
`confluent-local`) em containers Docker efêmeros durante os testes de
integração.

**Por que aqui.** Segue a convenção de testes do projeto
([memória: Go test structure]) — um `Test<Subject>` por unidade, com
subtestes tabulares via `t.Run` para caminho feliz e casos de erro.
Testes de integração usam a build tag `integration`
(`go test -tags=integration ./...`, separado via `make test-integration`
no Makefile), mantendo os testes unitários rápidos e sem Docker como
padrão (`make test`).

**Trade-offs.**
- ✅ Testcontainers testa contra o *comportamento real* do Postgres e
  do Kafka, não contra um mock do protocolo — pega bugs de SQL, de
  serialização, e (mais importante, ver 1.5) de protocolo Kafka que um
  fake nunca pegaria.
- ✅ Ambiente isolado por execução de teste — sem "banco de teste
  compartilhado" sujando estado entre execuções.
- ❌ Testes de integração são lentos (subir container Docker a cada
  execução) — por isso ficam atrás de uma build tag separada, não no
  `go test ./...` padrão.
- ❌ **Testcontainers valida o código contra um ambiente *efêmero e bem
  configurado*, não contra a configuração de infraestrutura real
  declarada em `docker-compose.yml`** — essa é exatamente a lacuna
  documentada no LEARNING-005: o Kafka do compose estava quebrado
  (`KAFKA_ADVERTISED_LISTENERS`) havia tempo, e nenhum teste com
  testcontainers (que usa `confluent-local` em KRaft, config própria)
  jamais teria pego isso. É uma limitação estrutural da técnica, não um
  uso incorreto dela.
- ⚠️ Requer Docker disponível na máquina/CI que roda
  `test-integration` — uma dependência de ambiente que testes unitários
  puros não têm.

### 1.9 Docker multi-stage build + Docker Compose

**O que é.** Cada serviço tem um `Dockerfile` próprio com build
multi-stage: uma etapa Go builder (`CGO_ENABLED=0 GOOS=linux
GOARCH=amd64 go build`) e uma imagem final mínima só com o binário
estático. O `docker-compose.yml` orquestra Postgres, Zookeeper, Kafka,
containers de migração, e os três serviços Go, com `healthcheck` no
Postgres e `depends_on` com condições (`service_healthy`,
`service_completed_successfully`) para sequenciar corretamente
migração → serviço.

**Trade-offs.**
- ✅ `CGO_ENABLED=0` + binário estático → imagem final pequena (base
  `scratch`/`alpine`), sem dependências de sistema em runtime.
- ✅ `depends_on` com condições evita a classe de bug "serviço subiu
  antes do banco estar pronto/migrado" — mas não evita bugs de
  *configuração* dentro dos próprios serviços declarados (ver 1.5).
- ❌ Orquestração via Compose é adequada para dev/local, mas não é o
  que rodaria em produção (Kubernetes, mencionado como próximo passo no
  README) — os healthchecks e `depends_on` do Compose não têm
  equivalente direto 1:1 em manifests k8s (que usariam readiness/liveness
  probes e, possivelmente, Jobs para migração).

---

## Parte 2 — Técnicas e padrões arquiteturais

### 2.1 Arquitetura Hexagonal / Ports & Adapters (ADR-002)

**O que é.** O núcleo do sistema (domínio + aplicação) define
*interfaces* (portas) para tudo que precisa do mundo externo
(persistência, mensageria); a infraestrutura implementa essas
interfaces (adaptadores). A dependência sempre aponta de fora para
dentro — domínio nunca importa `infrastructure`.

**Como aparece no código.** `application/port/output.UnitOfWork` e
`RepositoryProvider` são interfaces; `infrastructure/persistence/postgres`
as implementa. Os handlers de comando (`ReserveStockHandler` etc.)
recebem essas interfaces no construtor — nunca importam `pgx` ou SQL
diretamente.

**Trade-offs.**
- ✅ Domínio 100% testável sem banco/Kafka — `stockItem_test.go` testa
  `Reserve`/`Release`/`Replenish` como funções puras.
- ✅ Trocar Postgres por outro banco, ou Gin por outro framework HTTP,
  não toca domínio nem aplicação — só a implementação do adaptador.
- ❌ Mais indireção e mais arquivos por funcionalidade (interface +
  implementação + wiring) comparado a uma abordagem mais direta — custo
  de manutenção real para um time pequeno/projeto simples, compensado
  pela testabilidade e pela clareza de fronteiras à medida que o
  sistema cresce.
- ❌ Exige disciplina constante: é fácil "vazar" uma dependência de
  infraestrutura para dentro do domínio por atalho (ex: importar `pgx`
  num value object "só dessa vez") — não há enforcement automático, só
  revisão de código e a regra explícita na `ARCHITECTURE.md`.

### 2.2 DDD tático — Aggregate Root, Value Object, Domain Events, Repository

**Aggregate Root (`StockItem`).** Campos privados, mutação só via
método (`Reserve`, `Release`, `Replenish`), cada método valida
invariantes *antes* de mutar e retorna erro tipado se violado
(`domainErrors.ErrInsufficientStock` etc.). `isValid()` é uma checagem
de corrupção de estado interna, defensiva.

**Value Object (`SKU`).** Imutável, se autovalida na criação
(`NewSKU`), com `IsZero()` para representar "ausência" sem usar `nil`
ou string vazia soltas pelo código.

**Domain Events.** `StockReserved`, `StockReleased`,
`StockReplenished`, `LowStockDetected` — cada um representa um fato de
negócio já ocorrido (nome no passado), carregando só os dados
relevantes daquele fato. São acumulados no agregado (`addEvent`) e
extraídos pela aplicação via `PullEvents()` (que também limpa a lista —
side effect explícito e intencional, não um getter simples).

**Repository.** Interface de domínio (`domain/repository`) que fala a
língua do domínio (`FindBySKU(sku valueobject.SKU) (*StockItem, error)`),
implementada em infraestrutura traduzindo de/para linhas SQL via um
`StockMapper` dedicado.

**Trade-offs.**
- ✅ Invariantes de negócio ficam num único lugar, impossível de
  contornar — não existe caminho de código que reserve estoque negativo
  sem passar por `Reserve()`.
- ✅ Eventos de domínio nascem "de graça" junto com a mutação — não há
  risco de esquecer de emitir um evento em um dos vários pontos de
  entrada, porque a emissão está dentro do próprio método que muda o
  estado.
- ❌ Mais boilerplate que manipular structs com campos públicos
  diretamente — cada mutação precisa de um método explícito, cada
  leitura de um getter.
- ⚠️ `RestoreStockItem` (reconstrução a partir do banco) deliberadamente
  *não* revalida invariantes de criação nem emite eventos — correto
  (dado já persistido não deveria gerar side-effects de "criação"), mas
  é uma assimetria que exige cuidado: um bug de dados corrompidos no
  banco só seria pego por `isValid()` no próximo método de mutação, não
  no momento da reconstrução.

### 2.3 CQRS leve (Commands / Queries)

**O que é.** Separação de interface e caminho de código entre escrita
(`input.ReserveStockUseCase` etc., que passam pelo agregado + outbox) e
leitura (`input.StockQueries`, que faz `SELECT` direto para um DTO,
sem instanciar o agregado). **Não** é CQRS "pesado" — não há event
sourcing, nem store de leitura separado/replicado; é a mesma tabela
`stock_items` lida dos dois lados.

**Trade-offs.**
- ✅ Query não paga o custo de reconstruir um agregado inteiro (com
  suas validações e possíveis efeitos colaterais de reconstrução) só
  para exibir dados — projeta direto para o formato que a API precisa
  (`StockDTO`).
- ✅ Separação de interface deixa claro, só pelo nome do pacote/tipo,
  se um caso de uso muta ou só lê estado.
- ❌ Sem store de leitura separado, queries continuam competindo pelo
  mesmo banco/tabela que a escrita — não há o ganho de escalabilidade
  independente que um CQRS "cheio" com store de leitura dedicado
  traria. Isso é uma escolha proporcional ao estágio atual do projeto,
  não uma limitação acidental.

### 2.4 Unit of Work

**O que é.** Abstração que garante que múltiplas operações de
persistência (salvar o agregado + gravar eventos no outbox) aconteçam
na mesma transação de banco, com commit/rollback atômico.

**Como é implementado.** `unitOfWork.Do(ctx, fn)` abre uma `pgx.Tx`,
constrói um `repositoryProvider` que expõe `StockRepository()` e
`OutboxRepository()` *sobre a mesma transação*, executa `fn`, e faz
commit se `fn` não retornar erro (rollback caso contrário, inclusive em
`panic`, via `recover` + rollback + re-panic).

**Trade-offs.**
- ✅ Atomicidade garantida no nível certo: o handler de aplicação nunca
  precisa saber que existe uma transação por baixo, só que "tudo dentro
  de `Do` acontece ou nada acontece".
- ✅ `recover()` no `Do` evita que um `panic` dentro da função deixe uma
  transação pendurada sem rollback.
- ❌ O `RepositoryProvider` precisa ser reconstruído a cada chamada de
  `Do` (não é reutilizável entre transações) — um padrão correto mas
  que exige que todo novo repositório do bounded context seja também
  exposto na interface `RepositoryProvider`, criando um ponto de
  acoplamento entre "quantos repositórios existem" e "o que o UoW
  expõe".
- ⚠️ Um bounded context com muitos agregados diferentes tende a fazer
  esse `RepositoryProvider` crescer — hoje é pequeno (`StockRepository`
  + `OutboxRepository`), mas é um ponto a observar conforme o domínio
  cresce.

### 2.5 Transactional Outbox (ADR-003)

**O que é.** Em vez de publicar um evento diretamente no Kafka dentro
do mesmo fluxo que persiste o agregado (o que criaria uma janela de
inconsistência: e se o commit no banco funcionar mas a publicação no
Kafka falhar, ou vice-versa?), o evento é gravado numa tabela `outbox`
**na mesma transação** que persiste o agregado. Um processo separado
(o relay/cdc-connector) lê essa tabela depois e publica no Kafka de
forma assíncrona e desacoplada.

**Trade-offs (herdados da ADR-003, resumidos aqui).**
- ✅ Elimina o problema de "dual write" — a garantia de atomicidade do
  próprio banco relacional (ACID) é reaproveitada para garantir que
  "agregado salvo" e "evento vai ser publicado eventualmente" sejam a
  mesma coisa, sem precisar de um coordenador de transação distribuída
  (2PC) entre Postgres e Kafka.
- ✅ A aplicação nunca depende da disponibilidade do Kafka no caminho
  crítico de escrita — se o Kafka cair, `Reserve()` ainda funciona e o
  evento fica acumulado no outbox para quando o relay conseguir
  publicar.
- ❌ Introduz **latência de entrega** — o evento só chega ao Kafka no
  próximo ciclo de polling do relay (hoje até 2s), não é "publicado" no
  mesmo instante da transação. Não é um sistema near-real-time.
- ❌ Precisa de um processo separado (o relay) para fechar o ciclo —
  sem ele, a tabela outbox simplesmente acumula linhas não processadas
  para sempre (o que de fato aconteceu entre a Fase 0/1 e a
  implementação do cdc-connector, ver ADR-006).
- ⚠️ Garante *at-least-once* delivery, não *exactly-once* — se o relay
  publicar com sucesso mas cair antes de marcar a linha como
  processada, o mesmo evento será republicado no próximo ciclo (ver
  2.8).

### 2.6 Message Relay via Polling (cdc-connector, ADR-006)

**O que é.** O padrão que resolve a segunda metade do Transactional
Outbox: um processo dedicado que faz polling periódico da tabela
outbox de múltiplos bounded contexts, publica cada linha não
processada no Kafka, e marca como processada só após confirmação do
broker.

**Por que polling e não CDC baseado em log (Debezium/WAL).** Já
descartado desde a ADR-003 e reafirmado na ADR-006: CDC baseado em
log de replicação captura mudanças sem exigir escrita explícita na
tabela outbox e evita custo de polling repetido, mas exige Kafka
Connect, conectores Debezium e gerenciamento de replication slots —
complexidade operacional desproporcional ao estágio atual do projeto.
O contrato (`FetchUnprocessed`/`MarkProcessed`) foi desenhado
deliberadamente para que essa troca seja possível depois sem mudar
quem depende da interface.

**Design de isolamento.** `internal/relay` não conhece Postgres nem
Kafka diretamente — só duas interfaces pequenas, `OutboxReader` e
`Publisher`. Isso é o que permite testar a *lógica* do relay
(RunOnce, tratamento de erro parcial, batching) inteiramente com fakes,
e testar o *contrato* separadamente com Postgres+Kafka reais via
testcontainers.

**Trade-offs.**
- ✅ Simplicidade operacional — um processo Go simples, sem
  infraestrutura de CDC adicional.
- ✅ Falha isolada por linha/source — uma source fora do ar, ou uma
  linha com payload problemático, não trava o processamento das demais
  (cada falha só significa "tentar de novo no próximo ciclo").
- ✅ Adicionar uma nova source é configuração (`SOURCES=nome:banco,...`),
  não código.
- ❌ Latência limitada pelo `POLL_INTERVAL` (2s hoje) — não é
  near-real-time, ver 2.5.
- ❌ Ponto único de acoplamento operacional: um processo central com
  credenciais em *todo* banco de bounded context que ele drena —
  exceção deliberada ao princípio "cada serviço é dono exclusivo do seu
  banco" (ADR-002), justificada mas registrada como custo explícito na
  ADR-006 (mitigação recomendada: usuário de banco do cdc-connector
  restrito a `SELECT`/`UPDATE` só na tabela outbox, nunca acesso amplo
  às tabelas de domínio).
- ❌ Custo de polling constante no Postgres de cada source a cada
  ciclo, mesmo sem eventos novos.
- ⚠️ Se o cdc-connector cair, nenhum evento de nenhum serviço é
  entregue — mas nenhum dado é perdido (a tabela outbox continua
  acumulando; é uma falha de disponibilidade de entrega, não de
  durabilidade).

### 2.7 Shared Kernel (ADR-005)

**O que é.** Um módulo (`shared`) com código de infraestrutura
genuinamente comum a todo bounded context — `outbox` (contrato +
implementação Postgres), `events` (`DomainEvent`, `BaseEvent`),
`messaging/kafka` (producer), `postgres` (conexão) — para não duplicar
essa lógica em cada serviço.

**Onde termina o shared kernel.** Deliberadamente, `shared` não contém
nada de domínio específico (nenhum agregado, nenhuma regra de negócio de
um bounded context vaza para lá) — só infraestrutura reaproveitável e
os tipos mais genéricos de evento.

**Trade-offs.**
- ✅ Zero duplicação de código de outbox/Kafka/Postgres entre
  order-service, inventory-service e cdc-connector.
- ✅ Um bug corrigido no `shared/outbox` (ex.: uma query) se propaga
  para todo bounded context ao atualizar a dependência local — não
  precisa ser corrigido N vezes.
- ❌ Acoplamento de versão implícito: como o `replace` aponta para o
  caminho local (não uma versão fixada), uma mudança em `shared` afeta
  *todos* os serviços simultaneamente, sem possibilidade de um serviço
  ficar numa versão antiga do shared kernel enquanto outro já migrou —
  qualquer mudança "quebra" (ou muda o comportamento de) tudo de uma
  vez.
- ⚠️ Shared kernel é, por definição de DDD, um dos padrões de
  integração entre bounded contexts com *mais* acoplamento (comparado a
  Customer/Supplier ou Anticorruption Layer) — aceitável aqui porque o
  que é compartilhado é puramente técnico (outbox, mensageria), nunca
  modelo de domínio, mas é uma decisão que precisa continuar sendo
  vigiada para não "vazar" conceitos de negócio para dentro do
  `shared` conforme o sistema cresce (é literalmente o que a ADR-005
  registra ter quase acontecido e sido corrigido).

### 2.8 Entrega at-least-once / idempotência por convenção

**O que é.** O sistema, ponta a ponta (outbox → relay → Kafka →
consumidor), garante *at-least-once delivery*, nunca
*exactly-once*: é possível (embora raro) que um evento seja publicado
duas vezes — por exemplo, se `Publish` tiver sucesso mas o processo
cair antes de `MarkProcessed` rodar. Ele será encontrado de novo como
"não processado" no próximo ciclo e republicado.

**Onde a responsabilidade recai.** Nem o outbox, nem o relay,
deduplicam ativamente. A decisão (registrada na ADR-006, referenciando
ADR-003/004) é que **o consumidor** é responsável por lidar com
duplicatas — tipicamente via idempotência no processamento (ex.: usar
o ID do evento ou uma chave de negócio para ignorar reprocessamento).

**Trade-offs.**
- ✅ Mantém o relay e o outbox simples — deduplicação ativa exigiria
  estado adicional (ex.: uma tabela de dedup com TTL) só para cobrir
  uma janela de falha estreita.
- ✅ At-least-once é a garantia mais fácil de alcançar de forma
  correta com essa arquitetura (outbox + polling) — exactly-once
  distribuído é notoriamente difícil e, na prática, normalmente
  implementado como "at-least-once + idempotência no consumidor" de
  qualquer forma, então a escolha não abre mão de nada que seria fácil
  de ganhar de outro jeito.
- ❌ Todo consumidor futuro de eventos deste sistema precisa ser
  escrito com idempotência em mente desde o design — não é um detalhe
  que pode ser adicionado depois sem revisão; é um requisito
  arquitetural implícito herdado por qualquer novo assinante de tópico.

### 2.9 Convenção de nomes de tópico (`<bounded-context>.<evento>`)

**O que é.** Tópicos Kafka são nomeados dinamicamente pelo relay como
`source.Name + "." + row.Type` — ex.: `inventory-service.StockReserved`
(ADR-004).

**Trade-offs.**
- ✅ Convenção previsível e auto-descritiva — um consumidor sabe
  exatamente de qual bounded context e qual tipo de evento esperar só
  pelo nome do tópico, sem consultar documentação externa.
- ✅ Escala automaticamente com novos bounded contexts/eventos — nenhum
  registro central de tópicos precisa ser mantido manualmente; o nome
  emerge do próprio dado (`Source.Name`, `row.Type`).
- ❌ Um typo no `EventName()` de um evento de domínio cria um tópico
  novo silenciosamente (dado o `AllowAutoTopicCreation()`) em vez de
  falhar ruidosamente — não há validação de que `row.Type` pertence a
  um conjunto fechado de tipos conhecidos antes de publicar.
- ⚠️ Sem um registro central, é fácil dois eventos de bounded contexts
  diferentes colidirem em significado (não em nome, já que o prefixo
  evita isso) — mas não há nada que impeça, por exemplo, dois times
  nomeando eventos de forma inconsistente entre si (`StockReserved` vs
  `stock.reserved`) já que a convenção de *case*/estilo do `EventName()`
  não é validada automaticamente.

---

## Parte 3 — Trade-offs consolidados

| Escolha | Ganho principal | Custo principal |
|---|---|---|
| `pgx/v5` direto (não `database/sql`) | UoW nativo via `pgx.Tx` compartilhada | Acoplamento ao driver específico |
| `golang-migrate` em container separado | Migração auditável, sem race entre réplicas | Peça extra na orquestração |
| Kafka em Zookeeper (compose) vs KRaft (testes) | — | Divergência teste/dev que já escondeu 1 bug real |
| `franz-go` puro-Go | Build simples, sem cgo | Sem timeout de produção próprio (`ProduceSync` pode travar) |
| JSON sem schema registry | Simplicidade, debug legível | Sem garantia de compatibilidade de schema entre versões |
| Testcontainers para integração | Testa contra serviço real | Não substitui validar contra o `docker-compose.yml` real |
| Arquitetura Hexagonal | Domínio 100% testável e independente | Mais indireção/boilerplate |
| Transactional Outbox | Elimina dual-write | Latência de entrega (~2s) |
| Relay por polling (não CDC/WAL) | Simplicidade operacional | Ponto único de acoplamento (credenciais multi-banco) |
| Shared Kernel | Zero duplicação de infra comum | Acoplamento de versão total entre serviços |
| At-least-once (sem dedup) | Simplicidade no outbox/relay | Todo consumidor futuro precisa ser idempotente |

---

## Referências

- ADR-001 — Uso de PostgreSQL
- ADR-002 — Uso de Hexagonal Architecture
- ADR-003 — Uso de Transactional Outbox
- ADR-004 — Uso de Kafka
- ADR-005 — Shared Kernel entre bounded contexts
- ADR-006 — cdc-connector como relay de polling do Outbox
- LEARNING-003 — Unit of Work
- LEARNING-004 — Transactional Outbox
- LEARNING-005 — cdc-connector e o relay do Outbox
- [`ARCHITECTURE.md`](../../ARCHITECTURE.md)
