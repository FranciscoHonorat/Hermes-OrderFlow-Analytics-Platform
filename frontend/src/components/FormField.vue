<script setup lang="ts">
defineProps<{ label: string; error?: string }>()
const model = defineModel<string | number>()
</script>

<template>
  <label class="field">
    <span class="label">{{ label }}</span>
    <slot :model="model">
      <input v-model="model" />
    </slot>
    <span v-if="error" class="error">{{ error }}</span>
  </label>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-bottom: var(--space-3);
}

.label {
  font-size: 0.85rem;
  color: var(--color-text-muted);
}

input {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}

.error {
  color: var(--color-danger);
  font-size: 0.8rem;
}
</style>
