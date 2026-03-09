<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore, useWorkspaceStore, useChannelsStore } from '@/stores'
import { Avatar, Modal, Input, Button } from '@/components/ui'
import WorkspaceSwitcher from './WorkspaceSwitcher.vue'
import ChannelList from './ChannelList.vue'
import CallPanel from '@/components/calls/CallPanel.vue'
import WorkspaceSettingsModal from '@/components/workspace/WorkspaceSettingsModal.vue'
import UserSettingsModal from '@/components/workspace/UserSettingsModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()

const showCreateChannel = ref(false)
const showWorkspaceSettings = ref(false)
const showUserSettings = ref(false)

const currentUserRole = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  const members = workspaceStore.membersMap[wsId] ?? []
  return members.find(m => m.user_id === authStore.user?.id)?.role ?? null
})

const isAdmin = computed(() =>
  currentUserRole.value === 'owner' || currentUserRole.value === 'admin'
)
const newChannelName = ref('')
const newChannelType = ref<'text' | 'voice'>('text')
const newChannelPrivate = ref(false)
const newChannelCategoryId = ref<string | null>(null)
const creatingChannel = ref(false)

function normalizeChannelName(value: string) {
  return value.toLowerCase().replace(/\s+/g, '-').replace(/[^a-zа-яё0-9\-]/gi, '')
}

function onChannelNameInput(e: Event) {
  const input = e.target as HTMLInputElement
  const normalized = normalizeChannelName(input.value)
  newChannelName.value = normalized
  input.value = normalized
}

async function createChannel() {
  if (!newChannelName.value.trim() || !workspaceStore.currentWorkspaceId) return

  creatingChannel.value = true
  try {
    await channelsStore.createChannel({
      workspace_id: workspaceStore.currentWorkspaceId,
      name: newChannelName.value.trim(),
      type: newChannelType.value,
      is_private: newChannelPrivate.value,
      category_id: newChannelCategoryId.value,
    })
    showCreateChannel.value = false
    newChannelName.value = ''
    newChannelPrivate.value = false
    newChannelCategoryId.value = null
  } finally {
    creatingChannel.value = false
  }
}

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="w-64 flex flex-col bg-dark-900 border-r border-dark-800">
    <!-- Workspace switcher -->
    <WorkspaceSwitcher />

    <!-- Channels -->
    <div class="flex-1 overflow-y-auto">
      <ChannelList @create-channel="showCreateChannel = true" />
    </div>

    <!-- Панель активного голосового канала -->
    <CallPanel />

    <!-- Navigation -->
    <div class="p-3 border-t border-dark-800 space-y-1">
      <RouterLink
        to="/tasks"
        class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-dark-400 hover:text-dark-100 hover:bg-dark-800 transition-colors"
        active-class="bg-atlas-600/20 text-atlas-300"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
        </svg>
        Задачи
      </RouterLink>
      <template v-if="isAdmin">
        <button
          class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-dark-400 hover:text-dark-100 hover:bg-dark-800 transition-colors"
          @click="showCreateChannel = true"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Создать канал
        </button>
        <button
          class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-dark-400 hover:text-dark-100 hover:bg-dark-800 transition-colors"
          @click="showWorkspaceSettings = true"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Настройки воркспейса
        </button>
      </template>
    </div>

    <!-- User panel -->
    <div class="p-3 border-t border-dark-800 flex items-center gap-3">
      <Avatar
        v-if="authStore.user"
        :name="authStore.user.display_name"
        :src="authStore.user.avatar_url"
        size="sm"
        status="online"
      />
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-dark-100 truncate">
          {{ authStore.user?.display_name }}
        </p>
        <p class="text-xs text-dark-500 truncate">
          {{ authStore.user?.email }}
        </p>
      </div>
      <button
        class="btn-ghost p-2 rounded-lg text-dark-400 hover:text-dark-100"
        title="Настройки уведомлений"
        @click="showUserSettings = true"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
      </button>
      <button
        class="btn-ghost p-2 rounded-lg text-dark-400 hover:text-dark-100"
        title="Выйти"
        @click="logout"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
      </button>
    </div>

    <!-- Workspace settings modal -->
    <WorkspaceSettingsModal :open="showWorkspaceSettings" @close="showWorkspaceSettings = false" />

    <!-- User notification settings modal -->
    <UserSettingsModal :open="showUserSettings" @close="showUserSettings = false" />

    <!-- Create channel modal -->
    <Modal :open="showCreateChannel" title="Создать канал" @close="showCreateChannel = false">
      <form @submit.prevent="createChannel" class="space-y-4">
        <Input
          v-model="newChannelName"
          label="Название канала"
          placeholder="general"
          @input="onChannelNameInput"
        />

        <div class="space-y-2">
          <label class="block text-sm font-medium text-dark-300">Тип канала</label>
          <div class="flex gap-3">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="newChannelType"
                type="radio"
                value="text"
                class="w-4 h-4 text-atlas-600 bg-dark-800 border-dark-600"
              />
              <span class="text-sm text-dark-200">Текстовый</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="newChannelType"
                type="radio"
                value="voice"
                class="w-4 h-4 text-atlas-600 bg-dark-800 border-dark-600"
              />
              <span class="text-sm text-dark-200">Голосовой</span>
            </label>
          </div>
        </div>

        <!-- Категория -->
        <div v-if="channelsStore.categories.length > 0">
          <label class="block text-sm font-medium text-dark-300 mb-1.5">Категория</label>
          <select
            v-model="newChannelCategoryId"
            class="w-full px-3 py-2 bg-dark-900 border border-dark-700 rounded-lg text-dark-100 focus:border-atlas-500 focus:outline-none text-sm"
          >
            <option :value="null">Без категории</option>
            <option v-for="cat in channelsStore.categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <!-- Приватный канал -->
        <div
          class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
          :class="newChannelPrivate ? 'bg-atlas-600/10 border border-atlas-600/30' : 'bg-dark-800 border border-dark-700'"
          @click="newChannelPrivate = !newChannelPrivate"
        >
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 text-dark-400 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <div>
              <p class="text-sm font-medium text-dark-100">Приватный канал</p>
              <p class="text-xs text-dark-500">Только приглашённые участники</p>
            </div>
          </div>
          <div
            class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0"
            :class="newChannelPrivate ? 'bg-atlas-600' : 'bg-dark-600'"
          >
            <span
              class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
              :class="newChannelPrivate ? 'translate-x-[18px]' : 'translate-x-[3px]'"
            />
          </div>
        </div>

        <div class="flex gap-3 pt-2">
          <Button variant="secondary" class="flex-1" @click="showCreateChannel = false">
            Отмена
          </Button>
          <Button type="submit" :loading="creatingChannel" class="flex-1">
            Создать
          </Button>
        </div>
      </form>
    </Modal>
  </aside>
</template>

