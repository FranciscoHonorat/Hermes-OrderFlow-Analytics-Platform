<script setup lang="ts">
import { onMounted, ref } from 'vue'
import BaseButton from '@/components/BaseButton.vue'
import DataTable from '@/components/DataTable.vue'
import StatusPill from '@/components/StatusPill.vue'
import { activateUser, deactivateUser, listUsers, updateUserRole } from '@/api/users'
import { useToastStore } from '@/stores/toast'
import { extractErrorMessage } from '@/utils/errors'
import { formatDateTime } from '@/utils/format'
import type { UserDTO, UserRole } from '@/types/user'

const toast = useToastStore()
const users = ref<UserDTO[]>([])
const loading = ref(true)
const busyId = ref<string | null>(null)

const columns = [
  { key: 'email', label: 'Email' },
  { key: 'role', label: 'Papel' },
  { key: 'active', label: 'Status' },
  { key: 'created_at', label: 'Criado em' },
]

async function refresh() {
  loading.value = true
  try {
    const res = await listUsers()
    users.value = res.data
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    loading.value = false
  }
}

async function onChangeRole(user: UserDTO, role: UserRole) {
  if (role === user.role) return
  busyId.value = user.id
  try {
    await updateUserRole(user.id, { role })
    toast.push(`Papel de ${user.email} atualizado para ${role}`, 'success')
    await refresh()
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    busyId.value = null
  }
}

async function onToggleActive(user: UserDTO) {
  busyId.value = user.id
  try {
    if (user.active) {
      await deactivateUser(user.id)
      toast.push(`${user.email} desativado`, 'success')
    } else {
      await activateUser(user.id)
      toast.push(`${user.email} ativado`, 'success')
    }
    await refresh()
  } catch (err) {
    toast.push(extractErrorMessage(err), 'error')
  } finally {
    busyId.value = null
  }
}

onMounted(refresh)
</script>

<template>
  <div>
    <h1>Usuários</h1>

    <DataTable :items="users" :columns="columns" :loading="loading" empty-message="Nenhum usuário cadastrado.">
      <template #cell-active="{ item }">
        <StatusPill
          :label="item.active ? 'ativo' : 'inativo'"
          :tone="item.active ? 'success' : 'danger'"
        />
      </template>
      <template #cell-created_at="{ item }">{{ formatDateTime(item.created_at) }}</template>

      <template #actions="{ item }">
        <div class="actions">
          <select
            :value="item.role"
            :disabled="busyId === item.id"
            @change="onChangeRole(item, ($event.target as HTMLSelectElement).value as UserRole)"
          >
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
          <BaseButton
            variant="secondary"
            :disabled="busyId === item.id"
            @click="onToggleActive(item)"
          >
            {{ item.active ? 'Desativar' : 'Ativar' }}
          </BaseButton>
        </div>
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
.actions {
  display: flex;
  gap: var(--space-2);
  align-items: center;
}

select {
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius);
  border: 1px solid var(--color-border);
}
</style>
