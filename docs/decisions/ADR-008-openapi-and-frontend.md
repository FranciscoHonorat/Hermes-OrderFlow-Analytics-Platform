# ADR-008 — Documentação OpenAPI manual e frontend Vue 3/Vite

- Status: Accepted
- Date: 2026-09-19

## Context

O projeto tinha documentação de API apenas via Postman
(`docs/postman/`) e nenhum cliente além de curl/Postman para exercitar
os serviços. Duas necessidades surgiram: (1) uma documentação de API
mais padrão de mercado (OpenAPI/Swagger), navegável e independente de
uma ferramenta específica como o Postman; (2) um frontend que consuma
`order-service`, `inventory-service` e `user-service` como um
dashboard único, já que o `user-service` (ADR-007) agora fornece
autenticação real que um cliente pode usar.

## Decision

**OpenAPI**: três specs YAML escritos manualmente
(`docs/openapi/{order,inventory,user}-service.yaml`), um por bounded
context, mais um `docs/openapi/index.html` estático que carrega
Swagger UI via CDN e alterna entre os três. Não foi usado swaggo
(geração a partir de comentários nos handlers Go) — escolha do
usuário.

**Frontend**: Vue 3 + Vite + TypeScript em um novo diretório
`frontend/` na raiz do monorepo, com Vue Router, Pinia e Axios. Sem
biblioteca de componentes UI. Comunicação com os três backends via
proxy de desenvolvimento do próprio Vite, sem CORS nos serviços Go.

## Why YAML manual e não swaggo

swaggo gera o spec a partir de comentários `// @Summary`, `// @Param`
etc. anotados em cada handler — exigiria adicionar esses comentários
em todos os handlers já escritos de `order-service`/`inventory-service`
(que não foram construídos pensando nisso) e manter uma dependência
`swag`/`gin-swagger` adicional em cada `go.mod`. YAML manual é mais
trabalho para manter sincronizado com o código ao longo do tempo (risco
real: a spec pode divergir do comportamento real do handler), mas foi
a escolha explícita do usuário e não exige tocar nenhum código Go
existente. Mitigação parcial do risco de divergência: os specs
documentam explicitamente os comportamentos reais e não idealizados do
`order-service` (ex: `currency` sempre `"USD"`, `items` sempre vazio
na leitura) — foram escritos lendo o código, não reescrevendo o
contrato como "deveria" ser.

## Why proxy de dev do Vite e não CORS nos três serviços

Nenhum dos três serviços Go tem CORS configurado hoje (confirmado por
grep exaustivo no repo). Duas opções: adicionar `gin-contrib/cors` (ou
middleware equivalente) a `order-service`, `inventory-service` e
`user-service`, ou usar o proxy embutido do Vite
(`vite.config.ts` → `server.proxy`), que faz o dev server (não o
browser) falar com os três backends — o browser só vê same-origin. A
segunda opção não exige nenhuma mudança de código em nenhum dos três
serviços Go, o que a torna a escolha de menor superfície de mudança
para esta tarefa. O custo fica registrado abaixo: essa solução só
cobre desenvolvimento local; produção precisa de uma resposta própria
(reverse-proxy do frontend também em produção, ou CORS nos serviços —
ver Consequences).

## Why sem biblioteca de componentes UI

O escopo inicial é ~4 telas (login, dashboard, usuários, estoque,
pedidos) com tabelas e formulários simples — não uma superfície densa
o suficiente para justificar o peso e o sistema de temas de um kit
como Vuetify/PrimeVue/Element Plus. Um punhado de componentes
pequenos com CSS escopado, mais um arquivo de design tokens
compartilhado, é proporcional ao tamanho real do problema, seguindo o
mesmo princípio de "sem abstração além do necessário" já aplicado no
backend (ver ARCHITECTURE.md).

## Alternatives Considered

### swaggo/swag com gin-swagger

Geraria o spec automaticamente e o serviria em `/swagger/index.html`
em cada serviço, sem depender de um arquivo estático separado.
Descartado porque o usuário pediu explicitamente YAML manual.

### React ou outro framework para o frontend

Não considerado — o usuário pediu Vue.js com Vite explicitamente.

### Um serviço de API Gateway/BFF na frente dos três backends

Resolveria CORS e daria um único ponto de entrada para o frontend,
mas introduziria um serviço novo (mais um processo, mais uma peça de
infraestrutura) só para resolver um problema que o proxy de
desenvolvimento do Vite já resolve localmente sem custo. Fica como
possível evolução caso o frontend vá a produção.

## Trade-offs

### Benefícios

- OpenAPI é um formato padrão, consumível por qualquer ferramenta
  (Swagger UI, Postman, geradores de client), não só o ecossistema
  Postman já usado no projeto.
- Nenhuma mudança em código Go existente para viabilizar nem a
  documentação nem o frontend local.
- Frontend com footprint de dependências pequeno (`vue`, `vue-router`,
  `pinia`, `axios`), consistente com o resto do projeto.

### Custos

- Specs OpenAPI manuais podem divergir do código real ao longo do
  tempo — nenhuma validação automática (contract testing) as mantém
  sincronizadas hoje.
- O proxy do Vite só resolve o problema em desenvolvimento; uma build
  de produção do frontend (`npm run build`, servido estaticamente)
  volta a precisar de uma solução de CORS/reverse-proxy — deliberadamente
  adiado (ver Consequences).
- Sem componentes de UI prontos, formulários/tabelas mais complexos no
  futuro (paginação real, ordenação, validação rica) exigem mais
  código próprio do que usar um kit já pronto.

## Consequences

    Browser (localhost:5173, dev)
      │  same-origin
      ▼
    Vite dev server ──┬─▶ order-service:8080
                       ├─▶ inventory-service:8081
                       └─▶ user-service:8083

Dockerização do frontend fica deliberadamente adiada para uma etapa
futura. Quando retomada, a abordagem correta é o mesmo padrão de
reverse-proxy (agora via nginx dentro do container, apontando para os
hostnames do `docker-compose.yml` como `order-service:8080`), **não**
URLs absolutas (`http://localhost:8080`) embutidas no build — isso
reintroduziria cross-origin entre o browser e os serviços Go,
exatamente o problema que o proxy de desenvolvimento evita hoje.

## Security Considerations

**Atualizado em 2026-09-19**: `order-service` e `inventory-service`
não tinham autenticação própria quando este ADR foi escrito — o login
do frontend era só um gate de UX da aplicação Vue, sem nenhuma
garantia de segurança real nesses dois backends. Isso foi resolvido
(ver ADR-005 "Update" e ADR-007 "Update"): os dois agora exigem o
mesmo JWT do `user-service`. O `SecurityNoticeBanner` do frontend, que
avisava sobre essa lacuna, foi removido por não ser mais verdadeiro —
nenhuma outra mudança foi necessária no frontend, já que
`api/http.ts` sempre enviou o header `Authorization` para os três
serviços.

## Related Decisions

- ADR-002 — Uso de Hexagonal Architecture
- ADR-004 — Uso de Kafka
- ADR-007 — user-service: autenticação, hashing e bounded context dedicado
