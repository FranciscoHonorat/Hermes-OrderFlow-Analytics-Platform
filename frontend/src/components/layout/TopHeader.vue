<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'

const router = useRouter()
const auth = useAuthStore()
const cart = useCartStore()

const query = ref('')
const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

function onSearch() {
  const value = query.value.trim()
  if (!value) return

  if (UUID_RE.test(value)) {
    router.push(`/orders/${value}`)
  } else {
    router.push({ path: '/inventory', query: { sku: value } })
  }
}

function onLogout() {
  auth.logout()
  router.push('/login')
}

function onCartClick() {
  if (!auth.isAuthenticated) {
    router.push('/login')
    return
  }
  cart.toggle()
}
</script>

<template>
  <header class="top-header">
    <div class="row-main">
      <RouterLink to="/" class="brand">Hermes OrderFlow</RouterLink>

      <form class="search" @submit.prevent="onSearch">
        <input
          v-model="query"
          type="search"
          placeholder="Buscar pedido por ID ou item por SKU…"
        />
        <button type="submit" aria-label="Buscar">🔍</button>
      </form>
    </div>

    <div class="row-account">
      <template v-if="!auth.isAuthenticated">
        <RouterLink to="/register">Crie a sua conta</RouterLink>
        <RouterLink to="/login">Entre</RouterLink>
      </template>
      <template v-else>
        <span class="user-email">{{ auth.user?.email }}</span>
        <RouterLink to="/orders">Compras</RouterLink>
        <button class="link" @click="onLogout">Sair</button>
      </template>

      <button class="cart-button" @click="onCartClick" aria-label="Carrinho">
        🛒
        <span v-if="cart.count > 0" class="badge">{{ cart.count }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.top-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  padding: var(--space-3) var(--space-6);
}

.row-main {
  display: flex;
  align-items: center;
  gap: var(--space-5);
}

.brand {
  font-weight: 700;
  color: var(--color-text);
  text-decoration: none;
  white-space: nowrap;
}

.search {
  flex: 1;
  max-width: 640px;
  display: flex;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}

.search input {
  flex: 1;
  border: none;
  padding: var(--space-2) var(--space-3);
  outline: none;
}

.search button {
  border: none;
  background: var(--color-bg);
  padding: 0 var(--space-3);
  cursor: pointer;
}

.row-account {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: var(--space-4);
  margin-top: var(--space-2);
  font-size: 0.9rem;
}

.row-account a,
.row-account .link {
  color: var(--color-text);
  text-decoration: none;
  border: none;
  background: none;
  cursor: pointer;
  font-size: inherit;
  font-family: inherit;
}

.row-account a:hover,
.row-account .link:hover {
  color: var(--color-primary);
}

.user-email {
  color: var(--color-text-muted);
}

.cart-button {
  position: relative;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 1.1rem;
}

.badge {
  position: absolute;
  top: -6px;
  right: -10px;
  background: var(--color-primary);
  color: white;
  border-radius: 999px;
  font-size: 0.7rem;
  padding: 0 5px;
  line-height: 1.4;
}
</style>
