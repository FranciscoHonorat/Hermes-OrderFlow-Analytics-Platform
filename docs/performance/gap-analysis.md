# Gap analysis — Backend & Frontend Performance

Baseado nos dois checklists de boas práticas do roadmap.sh deixados na
raiz do projeto (`Screenshot ... Backend Performance Best Practices.png`
e `Screenshot ... Frontend Performance Best Practices.png`). Cada item
foi verificado contra o código real (não é uma opinião — cada linha
"Feito"/"Parcial" referencia o arquivo que prova isso), na mesma linha
de evidência já usada nas QA strategies e ADRs deste projeto.

Legenda: ✅ Feito · ⚠️ Parcial · ❌ Faltando · — Não se aplica hoje

---

## Parte A — Backend Performance

### Caching

| Item | Status | Evidência |
|---|---|---|
| Mecanismos de cache (HTTP/Server/Client/CDN) | ❌ | Nenhum header `Cache-Control`/`ETag` em nenhum handler; sem CDN |
| Padrão cache-aside/write-through/read-through | ❌ | `services/order-service/infrastructure/cache/redis/order_cache.go` existe (`Set`/`Get`/`Delete` genéricos) mas **nunca é instanciado em `cmd/server/main.go`** — código morto, não wired. Sem Redis no `docker-compose.yml`. |
| Estratégia de invalidação de cache | — | Não se aplica, já que não há cache real em uso |

### Optimize API Response

| Item | Status | Evidência |
|---|---|---|
| Limite de tamanho de payload | ❌ | Nenhum `MaxBytesReader`/limite de body em nenhum handler Gin |
| Compressão de resposta (gzip/brotli) | ❌ | Grep por `gzip`/`compress` em todo `services/`/`shared/`: zero ocorrências. Nenhum middleware `gin-contrib/gzip` |
| Paginação eficiente | ⚠️ | `order-service` tem `limit`/`offset` em `GET /orders` (`application/port/input/order_queries.go`); `inventory-service` (`GET /stock/low`) e `user-service` (`GET /users`) **não têm paginação nenhuma** — sempre retornam a tabela inteira |
| Minimizar processamento caro no servidor | ✅ | Handlers são finos (bind → use case → resposta), sem lógica pesada na camada HTTP |

### Databases

| Item | Status | Evidência |
|---|---|---|
| Connection pooling | ✅ | `shared/postgres/db.go`: `pgxpool` com `MaxConns=25`, `MinConns=5`, `MaxConnLifetime=30m` |
| Ajuste fino do pool | ✅ | Mesmos parâmetros acima — configurados, não são só o default |
| Índices eficientes | ⚠️ | Só a tabela `outbox` tem índice além da PK (`idx_outbox_unprocessed`, `WHERE processed_at IS NULL`, nos 3 serviços). **`orders.customer_id` e `orders.status` não têm índice** apesar de `GET /orders?customer_id=` filtrar por essa coluna — hoje é sequential scan. `users.email` tem índice só porque é `UNIQUE` (efeito colateral, não uma decisão de índice deliberada) |
| ORM bem ajustado | — | Não há ORM (SQL cru via `pgx/v5` em todo repositório, ver ADR-002) — item não se aplica |
| Lazy/eager loading, batch processing | ⚠️ | Sem ORM não há lazy/eager loading para "ajustar", mas o `cdc-connector` já faz *batch processing* de verdade (`POLL_BATCH_SIZE`, `FetchUnprocessed(ctx, limit)`) |
| Evitar `SELECT *` | ✅ | Todo repositório (`stock_repository.go`, `order_repository.go`, `user_repository.go`) lista colunas explicitamente |
| Desnormalização para leitura pesada | ❌ | Não avaliado — schema é normalizado por padrão em todo lugar |
| Otimizar/evitar JOINs desnecessários | ⚠️ | Efeito colateral de um bug conhecido, não de uma otimização: a query de leitura do `order-service` **não faz join nenhum** com a tabela de itens — por isso `GET /orders` sempre retorna `items: []` (documentado em `docs/openapi/order-service.yaml` e `OrderDetailView.vue`). É uma lacuna funcional, não uma otimização de performance |
| Limpeza/manutenção periódica (vacuum, etc.) | ❌ | Nenhuma rotina de manutenção agendada; tabela `outbox` cresce indefinidamente sem purge das linhas já processadas (risco já registrado na ADR-003, "Custos") |
| Slow-query logging | ❌ | Não configurado em nenhum serviço nem no `docker-compose.yml` do Postgres |
| Réplica de leitura | ❌ | Uma única instância `postgres:16-alpine` no compose, sem réplica |
| Sharding | ❌ | Não implementado (nem necessário no volume atual) |
| Ferramentas de profiling do banco | ❌ | Não usadas |

### Asynchronism

| Item | Status | Evidência |
|---|---|---|
| Offload de tarefas pesadas para background/filas | ✅ | O `cdc-connector` é exatamente isso — um worker separado drenando a tabela outbox (ADR-006) |
| Message broker para comunicação assíncrona | ✅ | Kafka, via `shared/messaging/kafka.Producer` + Transactional Outbox (ADR-003/ADR-004) |

### Load Balancing & Scaling

| Item | Status | Evidência |
|---|---|---|
| Scaling horizontal/vertical | ❌ | Cada serviço roda como um único container no `docker-compose.yml`, sem réplicas |
| Load balancing entre servidores | ❌ | Sem load balancer/reverse proxy na frente de nenhum serviço |

### Security

| Item | Status | Evidência |
|---|---|---|
| Dependências atualizadas | ⚠️ | `.github/dependabot.yml` cobre **só `gomod`** — os 4 módulos Go têm atualização automática semanal. **O `frontend/` (npm) não tem Dependabot configurado** |
| Autenticação/autorização | ✅ | JWT (HS256) em `user-service`, `order-service` e `inventory-service` via `shared/auth` — implementado nesta sessão (ver ADR-005 "Update", ADR-007 "Update") |
| Rate limiting / throttling | ❌ | Grep por `ratelimit`/`throttl`/`limiter` em todo o código: zero ocorrências. Qualquer cliente autenticado pode fazer requisições ilimitadas |
| Auditoria de segurança periódica | ⚠️ | Existe a skill `security-review` disponível no ambiente, mas não há um processo recorrente documentado nem executado neste projeto ainda |

### Code Optimization

| Item | Status | Evidência |
|---|---|---|
| Streaming de requests/responses grandes | ❌ | Todo handler usa `c.JSON`/`c.ShouldBindJSON`, sem streaming |
| Profiling de código | ❌ | Nenhum `pprof` habilitado em nenhum serviço |
| Otimização de algoritmos/estruturas de dados | — | Volume de dados do projeto é pequeno demais para isso ser um problema real hoje |
| Otimizar caminhos/endpoints críticos | ❌ | Nenhum endpoint foi identificado ou tratado como "crítico" com otimização dedicada |
| Linguagem compilada para partes críticas | ✅ | Go em todo o backend — já satisfeito por construção |
| Arquitetura SOA/Microservices | ✅ | 4 bounded contexts separados (order/inventory/user/cdc-connector), hexagonal architecture (ADR-002) |
| Timeouts de conexão + retry | ⚠️ | Conexão ao Postgres usa `context.WithTimeout(10s)` (`shared/postgres/db.go`); **não há nenhuma lógica de retry/backoff em lugar nenhum do código** (grep por `retry`/`backoff`: zero) — nem para reconexão de banco, nem para publicação no Kafka além do próximo ciclo de polling do relay |
| Batelar requisições similares | ✅ | O relay do `cdc-connector` já processa em lote (`POLL_BATCH_SIZE`) |

### Network

| Item | Status | Evidência |
|---|---|---|
| Hospedar próximo aos usuários | — | Sem deploy em produção ainda (só `docker-compose` local) |
| HTTP keep-alive | ⚠️ | Comportamento padrão do `net/http`/Gin (keep-alive habilitado por padrão no Go), mas nunca configurado/verificado explicitamente |
| CDN para assets estáticos | ❌ | Nenhum CDN configurado; o frontend nem tem build de produção servido ainda (Docker do frontend deliberadamente adiado, ver ADR-008) |
| Prefetch/preload de recursos | ❌ | Não implementado |

### Monitoring and Logging

| Item | Status | Evidência |
|---|---|---|
| Monitoramento/logging abrangente | ⚠️ | Todo serviço loga via `log.Printf`/`log.Println` para stdout — existe log, mas é só isso: sem correlação estruturada além do `X-Request-ID` (`middleware/request_id.go`, presente em todos os serviços mas não propagado para dentro dos logs) |
| Prometheus/Grafana/ELK | ❌ | Grep por `prometheus`/`grafana`/`opentelemetry`: zero ocorrências no código. O README já lista "Prometheus + Grafana" como parte da stack **planejada**, nunca implementada |
| Logging assíncrono | ❌ | `log.Printf` é síncrono (bloqueia a goroutine chamadora) |

### Performance Testing

| Item | Status | Evidência |
|---|---|---|
| Testes de performance/benchmark recorrentes | ❌ | Nenhum `go test -bench`, nenhum script de load testing (k6/vegeta/locust), nenhum job de CI para isso. `.github/workflows/go.yml` só builda e roda `go test ./...` (e, aliás, esse workflow está quebrado — roda da raiz do repo, que não funciona com o `go.work` multi-módulo; problema já documentado em `docs/QA/order-service-test-strategy.md`, nunca corrigido) |

---

## Parte B — Frontend Performance

### High Priority

| Item | Status | Evidência |
|---|---|---|
| Page weight < 500KB (ideal) / 1500KB | ✅ | `npm run build` → `dist/` = **276KB** total (sem gzip), ~61KB gzipped só o JS principal — bem abaixo do ideal |
| Page load time < 3s | — | Não medido formalmente (sem Lighthouse/PageSpeed rodado); o tamanho pequeno do bundle é um bom indício, mas não é medição |
| GZIP/Brotli habilitado | ❌ | Não há servidor de produção configurado ainda para o frontend (Docker adiado, ADR-008) — quando houver, precisa habilitar compressão no nginx |
| Minimizar requisições HTTP | ✅ | Code-splitting por rota (`router/index.ts`, `import()` dinâmico por view) — poucos chunks, carregados sob demanda |
| Comprimir imagens / poucas imagens | — | O app não tem nenhuma imagem hoje (removi os placeholders `vite.svg`/`vue.svg` do template) |
| Cache headers HTTP configurados | ❌ | Mesma causa do gzip: sem servidor de produção ainda |
| TTFB < 1.3s | — | Não medido (só dev server local) |
| JS não-bloqueante (`async`/`defer`) | ✅ | `dist/index.html` usa `<script type="module" crossorigin>` — módulos ES já são deferred por especificação |
| JS minificado | ✅ | Build de produção via Vite/esbuild minifica por padrão (`index-*.js` sem espaçamento) |
| CSS minificado | ✅ | Mesmo pipeline — confirmado no `dist/assets/*.css` gerado |
| Critical CSS inline (above the fold) | ❌ | Não implementado — precisaria de um plugin dedicado (`vite-plugin-critical` ou similar); todo CSS vai por `<link rel="stylesheet">` bloqueante |
| CSS não bloqueante | ❌ | Consequência direta do item acima — `<link rel="stylesheet">` padrão bloqueia render |
| HTTPS | ❌ | Só HTTP local (`localhost:5173`/portas dos serviços); sem TLS configurado em lugar nenhum |
| Formato de imagem apropriado | — | N/A, sem imagens |
| Evitar arquivos inacessíveis (404) | — | Não auditado formalmente, mas a superfície é pequena o suficiente para não ser um risco real hoje |
| Servir do mesmo protocolo | ✅ | Tudo HTTP hoje, sem mixed content |
| Minimizar iframes | ✅ | Zero iframes no app |
| Evitar CSS inline/embutido | ✅ | Vue SFC `<style scoped>` é compilado para CSS real em arquivo, não inline |
| Analisar complexidade das folhas de estilo | ❌ | Não analisado |

### Medium Priority

| Item | Status | Evidência |
|---|---|---|
| HTML minificado | ⚠️ | `dist/index.html` **não está minificado** — Vite mantém indentação/quebras de linha por padrão (confirmado lendo o arquivo gerado). Impacto real é desprezível dado o tamanho do arquivo (poucas linhas), mas o item, tecnicamente, não está satisfeito |
| CDN | ❌ | Não usado |
| Imagem vetorial > bitmap | — | N/A, sem imagens |
| width/height em imagens | — | N/A |
| Evitar Base64 | — | N/A |
| Lazy load de imagens fora da viewport | — | N/A |
| Servir imagem no tamanho de exibição | — | N/A |
| Evitar múltiplos `<script>` inline | ✅ | Nenhum script inline, tudo via módulos importados |
| Dependências atualizadas | ⚠️ | Versões atuais são recentes (Vue 3.5, Vite 8, TS 6) mas **sem Dependabot cobrindo `npm`** (só `gomod, ver Security acima) — atualização é manual |
| Checar problemas de performance no JS | ❌ | Nenhum profiling feito (Chrome DevTools/Lighthouse) |
| Service Workers | ❌ | Não implementado — sem PWA, sem cache offline |
| Tamanho de cookie < 4096 bytes | — | O app usa `localStorage` para o token JWT, não cookies — item não se aplica |
| Menos de 20 cookies | — | Mesmo motivo acima |

### Low Priority

| Item | Status | Evidência |
|---|---|---|
| Pré-carregar URLs | ❌ | Não implementado |
| Concatenar CSS em um arquivo | ⚠️ | Vite gera **um CSS por chunk/rota** (`UsersView-*.css`, `OrdersView-*.css`, etc.) em vez de um único arquivo — intencional para code-splitting, mas é o oposto literal deste item do checklist; trade-off consciente, não descuido |
| Remover CSS não usado | ❌ | Sem PurgeCSS/ferramenta equivalente (não há Tailwind, então o CSS já é escrito à mão e tende a ser pequeno, mas nada audita sobras) |
| Fontes em WOFF2 | — | O app não carrega nenhuma web font — usa só a pilha de fontes do sistema (`tokens.css`: `-apple-system, BlinkMacSystemFont, 'Segoe UI', ...`) — zero custo de fonte, o que já é melhor que a maioria dos itens desta seção |
| Preconnect para fontes | — | N/A, mesmo motivo |
| Fonte web < 300kb | — | N/A |
| Evitar FOIT/texto invisível | — | N/A |
| Vigiar tamanho das dependências | ✅ | Só 4 dependências de runtime (`vue`, `vue-router`, `pinia`, `axios`), nenhuma delas pesada — nunca rodei o Bundlephobia formalmente, mas o bundle final de 161KB não-gzipado já é evidência indireta |

---

## Backlog priorizado

O que teria mais impacto real para este projeto, na ordem em que eu
atacaria:

1. **Índice em `orders.customer_id`** — é o único índice faltando que
   já tem um caso de uso real e ativo (`GET /orders?customer_id=`
   fazendo sequential scan hoje). Migration pequena, baixo risco.
2. **Rate limiting básico** — hoje um cliente autenticado pode martelar
   qualquer endpoint sem limite. Um middleware simples (`golang.org/x/time/rate`
   por IP ou por usuário) nos três serviços fecha uma lacuna de
   segurança/estabilidade real, não só de performance.
3. **Compressão de resposta HTTP** (`gin-contrib/gzip`) nos três
   serviços — mudança pequena, ganho direto em toda resposta JSON.
4. **Paginação em `GET /stock/low` e `GET /users`** — hoje escala mal
   por construção; ainda não dói porque o volume de dados é pequeno,
   mas é uma bomba-relógio conhecida (mesmo padrão do `limit`/`offset`
   que `order-service` já tem).
5. **Remover ou de fato usar `order_cache.go`** — hoje é código morto
   com uma dependência (`go-redis`) no `go.mod` que nunca é exercitada.
   Ou vira cache de verdade (com Redis no compose) ou é deletado —
   estado atual é o pior dos dois mundos.
6. **Retry/backoff em publicações Kafka e reconexão Postgres** — hoje
   uma falha transitória de rede não tem nenhuma tentativa automática
   de recuperação além do próximo ciclo do relay (que só cobre o
   outbox, não conexões diretas).
7. **CSS crítico inline + HTML minificado** no frontend — ganho
   pequeno dado o tamanho já enxuto do bundle, mas barato de
   implementar via plugin do Vite quando o Docker do frontend for
   retomado (ADR-008).
8. **Observabilidade real (Prometheus/Grafana)** — já está no roadmap
   do README como planejado; hoje o único sinal de saúde de cada
   serviço é `/health` e log síncrono em stdout.
9. **CI corrigido** — `.github/workflows/go.yml` roda `go test ./...`
   da raiz, que falha com o workspace multi-módulo (bug já conhecido,
   nunca corrigido). Sem isso, nenhum teste de performance futuro
   (benchmark, load test) tem onde rodar de forma confiável.
10. **Dependabot para o `frontend/`** — hoje só os módulos Go recebem
    atualização automática de dependência.

## O que já está genuinely bem coberto

Vale reconhecer, não só listar o que falta: connection pooling do
Postgres, autenticação JWT cross-service, arquitetura de microsserviços
com mensageria assíncrona real (Kafka + Transactional Outbox), uso de
Go (linguagem compilada) e um bundle de frontend já bem enxuto por
construção (sem UI kit pesado, sem web fonts, poucas dependências) são
itens do checklist que este projeto já cumpre de verdade, não por
acaso — são decisões arquiteturais já tomadas e documentadas nas ADRs
002–007.

## Referências

- `docs/decisions/ADR-002-use-hexagonal-architecture.md`
- `docs/decisions/ADR-003-use-transactional-outbox.md`
- `docs/decisions/ADR-004-use-kafka.md`
- `docs/decisions/ADR-005-shared-kernel.md`
- `docs/decisions/ADR-007-user-service-authentication.md`
- `docs/QA/order-service-test-strategy.md` (bug do CI já documentado)
- `README.md` (Prometheus/Grafana como stack planejada)
