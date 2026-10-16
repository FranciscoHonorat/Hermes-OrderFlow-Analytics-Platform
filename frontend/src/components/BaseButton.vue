<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'danger'
    type?: 'button' | 'submit'
    loading?: boolean
    disabled?: boolean
  }>(),
  {
    variant: 'primary',
    type: 'button',
    loading: false,
    disabled: false,
  },
)
</script>

<template>
  <button :type="type" class="btn" :class="variant" :disabled="disabled || loading">
    <span v-if="loading">…</span>
    <slot v-else />
  </button>
</template>

<style scoped>
.btn {
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 600;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.primary {
  background: var(--color-primary);
  color: white;
}
.primary:not(:disabled):hover {
  background: var(--color-primary-hover);
}

.secondary {
  background: var(--color-surface);
  border-color: var(--color-border);
  color: var(--color-text);
}
.secondary:not(:disabled):hover {
  border-color: var(--color-primary);
}

.danger {
  background: var(--color-danger);
  color: white;
}
.danger:not(:disabled):hover {
  background: var(--color-danger-hover);
}
</style>
