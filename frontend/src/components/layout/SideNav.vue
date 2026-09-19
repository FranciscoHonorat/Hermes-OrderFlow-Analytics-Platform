<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <nav class="side-nav">
    <RouterLink to="/" exact-active-class="active">Dashboard</RouterLink>
    <RouterLink to="/orders" active-class="active">Pedidos</RouterLink>
    <RouterLink to="/inventory" active-class="active">Estoque</RouterLink>
    <RouterLink v-if="auth.isAdmin" to="/users" active-class="active">Usuários</RouterLink>
  </nav>
</template>

<style scoped>
.side-nav {
  display: flex;
  flex-direction: column;
  width: 200px;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  padding: var(--space-4);
  gap: var(--space-2);
}

.side-nav a {
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius);
  color: var(--color-text);
  text-decoration: none;
}

.side-nav a.active {
  background: var(--color-primary);
  color: white;
}
</style>
