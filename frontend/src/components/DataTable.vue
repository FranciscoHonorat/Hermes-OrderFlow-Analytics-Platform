<script setup lang="ts" generic="T extends Record<string, unknown>">
defineProps<{
  items: T[]
  columns: { key: string; label: string }[]
  loading?: boolean
  emptyMessage?: string
}>()
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.key">{{ col.label }}</th>
          <th v-if="$slots.actions">Ações</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td :colspan="columns.length + ($slots.actions ? 1 : 0)">Carregando…</td>
        </tr>
        <tr v-else-if="items.length === 0">
          <td :colspan="columns.length + ($slots.actions ? 1 : 0)">
            {{ emptyMessage ?? 'Nada para exibir.' }}
          </td>
        </tr>
        <tr v-for="(item, index) in items" :key="index" v-else>
          <td v-for="col in columns" :key="col.key">
            <slot :name="`cell-${col.key}`" :item="item">{{ item[col.key] }}</slot>
          </td>
          <td v-if="$slots.actions">
            <slot name="actions" :item="item" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}

th {
  color: var(--color-text-muted);
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

tbody tr:last-child td {
  border-bottom: none;
}
</style>
