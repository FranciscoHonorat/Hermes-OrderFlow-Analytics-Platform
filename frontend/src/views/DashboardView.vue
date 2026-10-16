<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listOrders } from '@/api/orders'
import { listLowStock } from '@/api/inventory'
import { listUsers } from '@/api/users'
import { useAuthStore } from '@/stores/auth'
import type { OrderDTO } from '@/types/order'
import type { StockDTO } from '@/types/inventory'

const auth = useAuthStore()

const recentOrders = ref<OrderDTO[]>([])
const lowStock = ref<StockDTO[]>([])
const userCount = ref<number | null>(null)
const loading = ref(true)

onMounted(async () => {
  const tasks: Promise<void>[] = [
    listOrders({ limit: 5 }).then((res) => {
      recentOrders.value = res.data
    }),
    listLowStock().then((res) => {
      lowStock.value = res.data
    }),
  ]

  if (auth.isAdmin) {
    tasks.push(
      listUsers()
        .then((res) => {
          userCount.value = res.data.length
        })
        .catch(() => {
          userCount.value = null
        }),
    )
  }

  await Promise.allSettled(tasks)
  loading.value = false
})
</script>

<template>
  <div>
    <h1>Dashboard</h1>

    <div class="cards">
      <div class="card">
        <h3>Pedidos recentes</h3>
        <p class="metric">{{ loading ? '…' : recentOrders.length }}</p>
      </div>
      <div class="card">
        <h3>Itens com estoque baixo</h3>
        <p class="metric">{{ loading ? '…' : lowStock.length }}</p>
      </div>
      <div class="card" v-if="auth.isAdmin">
        <h3>Usuários cadastrados</h3>
        <p class="metric">{{ loading ? '…' : (userCount ?? '—') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
  margin-top: var(--space-5);
}

.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--space-5);
}

.card h3 {
  margin: 0 0 var(--space-2);
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.metric {
  font-size: 2rem;
  font-weight: 700;
  margin: 0;
}
</style>
