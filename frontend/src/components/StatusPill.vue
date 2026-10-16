<script setup lang="ts">
const props = defineProps<{ label: string; tone?: 'success' | 'warning' | 'danger' | 'neutral' }>()

const toneClass = props.tone ?? 'neutral'
</script>

<template>
  <span class="pill" :class="toneClass">{{ label }}</span>
</template>

<style scoped>
.pill {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 600;
}

.success {
  background: #e3f6e9;
  color: var(--color-success);
}
.warning {
  background: #fbf0dd;
  color: var(--color-warning);
}
.danger {
  background: #fbe6e9;
  color: var(--color-danger);
}
.neutral {
  background: var(--color-border);
  color: var(--color-text-muted);
}
</style>
