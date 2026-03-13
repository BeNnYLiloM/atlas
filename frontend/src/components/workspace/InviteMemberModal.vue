<script setup lang="ts">
import { ref } from 'vue'
import { usersApi, workspacesApi } from '@/api'
import { useWorkspaceStore } from '@/stores'
import { Modal, Input, Button, Select } from '@/components/ui'
import type { User } from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const workspaceStore = useWorkspaceStore()

const email = ref('')
const role = ref<'admin' | 'member'>('member')
const foundUser = ref<User | null>(null)
const searching = ref(false)
const inviting = ref(false)
const error = ref('')
const success = ref('')

async function searchUser() {
  if (!email.value.trim()) return

  searching.value = true
  error.value = ''
  foundUser.value = null

  try {
    foundUser.value = await usersApi.searchByEmail(email.value.trim())
  } catch {
    error.value = 'Пользователь не найден'
  } finally {
    searching.value = false
  }
}

async function inviteUser() {
  if (!foundUser.value || !workspaceStore.currentWorkspaceId) return

  inviting.value = true
  error.value = ''

  try {
    await workspacesApi.addMember(
      workspaceStore.currentWorkspaceId,
      foundUser.value.id,
      role.value
    )
    success.value = `${foundUser.value.display_name} добавлен в воркспейс`
    foundUser.value = null
    email.value = ''
    
    // Закрываем через 2 секунды
    setTimeout(() => {
      success.value = ''
      emit('close')
    }, 2000)
  } catch {
    error.value = 'Не удалось добавить пользователя'
  } finally {
    inviting.value = false
  }
}

function onClose() {
  email.value = ''
  foundUser.value = null
  error.value = ''
  success.value = ''
  emit('close')
}

const roles = [
  { value: 'admin', label: 'Администратор' },
  { value: 'member', label: 'Участник' },
]
</script>

<template>
  <Modal
    :open="props.open"
    title="Пригласить участника"
    @close="onClose"
  >
    <div class="space-y-4">
      <!-- Success message -->
      <div
        v-if="success"
        class="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg"
      >
        <p class="text-sm text-emerald-400">
          {{ success }}
        </p>
      </div>

      <!-- Search form -->
      <div v-else>
        <div class="flex gap-2">
          <Input
            v-model="email"
            type="email"
            placeholder="email@example.com"
            class="flex-1"
            @keyup.enter="searchUser"
          />
          <Button
            :loading="searching"
            :disabled="!email.trim()"
            @click="searchUser"
          >
            Найти
          </Button>
        </div>

        <!-- Error -->
        <p
          v-if="error"
          class="mt-2 text-sm text-red-400"
        >
          {{ error }}
        </p>

        <!-- Found user -->
        <div
          v-if="foundUser"
          class="mt-4 p-4 bg-elevated rounded-lg space-y-4"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-accent flex items-center justify-center text-white font-semibold">
              {{ foundUser.display_name[0].toUpperCase() }}
            </div>
            <div>
              <p class="font-medium text-primary">
                {{ foundUser.display_name }}
              </p>
              <p class="text-sm text-muted">
                {{ foundUser.email }}
              </p>
            </div>
          </div>

          <!-- Role select -->
          <div>
            <label class="block text-sm font-medium text-tertiary mb-2">Роль</label>
            <Select
              v-model="role"
              :options="roles"
            />
          </div>

          <Button
            :loading="inviting"
            class="w-full"
            @click="inviteUser"
          >
            Пригласить
          </Button>
        </div>
      </div>
    </div>
  </Modal>
</template>


