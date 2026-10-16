# Hermes OrderFlow — Frontend

Dashboard em Vue 3 + Vite + TypeScript consumindo `order-service`,
`inventory-service` e `user-service`. Ver
[`ADR-008`](../docs/decisions/ADR-008-openapi-and-frontend.md) para as
decisões de arquitetura por trás deste app.

## Rodando localmente

Pré-requisitos: os três backends no ar (via `docker compose up` na
raiz do repo, ou `go run` individual) nas portas padrão:

- `order-service` → `:8080`
- `inventory-service` → `:8081`
- `user-service` → `:8083`

```bash
npm install
npm run dev
```

Abre em `http://localhost:5173`. Não é necessário CORS nos serviços
Go: o Vite (`vite.config.ts`) faz proxy de `/api/order`,
`/api/inventory` e `/api/user` para as portas acima — o browser só
fala com o dev server, same-origin.

Login: crie uma conta em `/register`, ou use o admin de bootstrap
definido em `BOOTSTRAP_ADMIN_EMAIL`/`BOOTSTRAP_ADMIN_PASSWORD` no
`docker-compose.yml` do `user-service`.

## Layout do header

O header (logo + barra de busca + linha de conta/carrinho, inspirado
livremente na estrutura de um marketplace — sem copiar a identidade
visual de nenhum) aparece tanto nas páginas públicas (`/login`,
`/register`) quanto dentro do dashboard autenticado:

- Busca: aceita um UUID (navega para o pedido correspondente) ou texto
  livre (navega para Estoque com o SKU pré-preenchido na busca).
- Linha de conta: "Crie a sua conta" / "Entre" quando deslogado;
  email do usuário + "Compras" (`/orders`) + "Sair" quando logado.
- Carrinho: monta um pedido (customer_id + itens) e chama
  `POST /orders` ao finalizar — é o mesmo fluxo que antes vivia num
  modal dentro de `/orders`, agora acessível de qualquer tela via
  `stores/cart.ts`.

## Build de produção

```bash
npm run build
```

Gera `dist/` (já ignorado no `.gitignore` da raiz). Dockerização deste
frontend (nginx servindo `dist/` com reverse-proxy para os serviços)
está deliberadamente adiada — ver a seção "Consequences" da ADR-008
para o desenho já pensado para quando for retomado.

## Estrutura

```
src/
  api/       clientes HTTP por serviço (axios) + interceptors de auth
  stores/    Pinia: auth (JWT + usuário, persistido em localStorage), toast
  router/    rotas + guards (requiresAuth/requiresAdmin)
  types/     interfaces TS espelhando os DTOs de cada backend
  components/ componentes reutilizáveis pequenos, sem UI kit
  views/     telas: login, dashboard, usuários, estoque, pedidos
```

## Limitações conhecidas do backend refletidas na UI

- `order-service` sempre retorna `currency: "USD"` e `items: []` na
  leitura de pedidos — a UI documenta isso explicitamente em vez de
  esconder o comportamento real (ver `OrderDetailView.vue`).

`order-service` e `inventory-service` agora exigem o mesmo JWT do
`user-service` em todas as rotas (exceto `/health`) — ver ADR-005
"Update" (`shared/auth`) e ADR-008. Não é mais necessário nenhum aviso
de segurança na UI para isso; o `api/http.ts` já enviava o header
`Authorization` para os três serviços desde que o frontend foi
construído, então nenhuma mudança foi necessária aqui além de remover
o aviso que ficou obsoleto.
