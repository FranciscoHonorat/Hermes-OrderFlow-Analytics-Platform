<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseButton from '@/components/BaseButton.vue'
import FormField from '@/components/FormField.vue'
import { register } from '@/api/users'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'

const router = useRouter()
const toast = useToastStore()

const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function onSubmit() {
  loading.value = true
  error.value = ''
  try {
    await register({ email: email.value, password: password.value })
    toast.push('Conta criada com sucesso — faça login.', 'success')
    router.push('/login')
  } catch (err) {
    error.value = extractErrorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="register-page">
    <form class="register-card" @submit.prevent="onSubmit">
      <h1>Criar conta</h1>
      <p class="subtitle">Novas contas começam com o papel "user"</p>

      <FormField label="Email" v-model="email">
        <input v-model="email" type="email" required autocomplete="username" />
      </FormField>
      <FormField label="Senha" v-model="password">
        <input v-model="password" type="password" required minlength="6" autocomplete="new-password" />
      </FormField>

      <p v-if="error" class="error">{{ error }}</p>

      <BaseButton type="submit" :loading="loading">Criar conta</BaseButton>

      <p class="footer-link">
        Já tem conta? <RouterLink to="/login">Entrar</RouterLink>
      </p>
    </form>
  </div>
</template>

<style scoped>
.register-page {
  min-height: calc(100vh - 96px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.register-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: var(--space-6);
  width: min(360px, 90vw);
}

h1 {
  margin: 0 0 var(--space-1);
  font-size: 1.3rem;
}

.subtitle {
  color: var(--color-text-muted);
  margin: 0 0 var(--space-5);
  font-size: 0.9rem;
}

.error {
  color: var(--color-danger);
  font-size: 0.85rem;
}

.footer-link {
  margin-top: var(--space-4);
  font-size: 0.85rem;
  text-align: center;
}
</style>
