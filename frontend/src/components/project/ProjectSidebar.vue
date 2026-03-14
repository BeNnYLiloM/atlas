<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useChannelsStore, useProjectsStore, useWorkspaceStore, useAuthStore } from '@/stores'
import { Modal, Input, Button } from '@/components/ui'
import ProjectSettingsModal from './ProjectSettingsModal.vue'

const router = useRouter()
const route = useRoute()
const channelsStore = useChannelsStore()
const projectsStore = useProjectsStore()
const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()

const showSettings = ref(false)
const showCreateChannel = ref(false)
const newChannelName = ref('')
const newChannelType = ref<'text' | 'voice'>('text')
const newChannelPrivate = ref(false)
const newChannelCategoryId = ref<string | null>(null)
const creatingChannel = ref(false)

const showCreateCategory = ref(false)
const newCategoryName = ref('')
const newCategoryPrivate = ref(false)
const creatingCategory = ref(false)

const collapsedCategories = ref<Record<string, boolean>>({})
function toggleCategory(id: string) {
  collapsedCategories.value[id] = !collapsedCategories.value[id]
}

const project = computed(() => projectsStore.currentProject)

type ChannelGroup = { id: string | null; name: string | null; channels: typeof channelsStore.channels }

const groupedTextChannels = computed<ChannelGroup[]>(() => {
  const cats = channelsStore.categories
  const uncategorized = channelsStore.channels.filter(c => c.type === 'text' && !c.category_id)
  const groups: ChannelGroup[] = []
  if (uncategorized.length > 0 || cats.length === 0) {
    groups.push({ id: null, name: null, channels: uncategorized })
  }
  for (const cat of cats) {
    const catChannels = channelsStore.channels.filter(c => c.type === 'text' && c.category_id === cat.id)
    groups.push({ id: cat.id, name: cat.name, channels: catChannels })
  }
  return groups
})

const groupedVoiceChannels = computed<ChannelGroup[]>(() => {
  const cats = channelsStore.categories
  const uncategorized = channelsStore.channels.filter(c => c.type === 'voice' && !c.category_id)
  const groups: ChannelGroup[] = []
  if (uncategorized.length > 0) {
    groups.push({ id: null, name: null, channels: uncategorized })
  }
  for (const cat of cats) {
    const catChannels = channelsStore.channels.filter(c => c.type === 'voice' && c.category_id === cat.id)
    if (catChannels.length > 0) {
      groups.push({ id: cat.id, name: cat.name, channels: catChannels })
    }
  }
  return groups
})

const hasAnyChannels = computed(() => channelsStore.channels.length > 0)

const currentMember = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  return (workspaceStore.membersMap[wsId] ?? []).find(m => m.user_id === authStore.user?.id) ?? null
})

const canManage = computed(() => {
  if (!project.value) return false
  if (currentMember.value?.role === 'owner' || currentMember.value?.role === 'admin') return true
  const pm = projectsStore.currentMembers.find(m => m.user_id === authStore.user?.id)
  return pm?.is_lead ?? false
})

function normalizeChannelName(value: string) {
  return value.toLowerCase().replace(/\s+/g, '-').replace(/[^a-zа-яё0-9-]/gi, '')
}

function onChannelNameInput(e: Event) {
  const input = e.target as HTMLInputElement
  const normalized = normalizeChannelName(input.value)
  newChannelName.value = normalized
  input.value = normalized
}

async function createChannel() {
  if (!newChannelName.value.trim() || !workspaceStore.currentWorkspaceId || !project.value) return
  creatingChannel.value = true
  try {
    const channel = await channelsStore.createChannel({
      workspace_id: workspaceStore.currentWorkspaceId,
      name: newChannelName.value.trim(),
      type: newChannelType.value,
      is_private: newChannelPrivate.value,
      category_id: newChannelCategoryId.value,
      project_id: project.value.id,
    })
    showCreateChannel.value = false
    newChannelName.value = ''
    newChannelPrivate.value = false
    newChannelType.value = 'text'
    newChannelCategoryId.value = null
    if (channel.type === 'text') selectChannel(channel.id)
  } finally {
    creatingChannel.value = false
  }
}

async function createCategory() {
  if (!newCategoryName.value.trim() || !workspaceStore.currentWorkspaceId || !project.value) return
  creatingCategory.value = true
  try {
    await channelsStore.createCategory(workspaceStore.currentWorkspaceId, {
      name: newCategoryName.value.trim(),
      is_private: newCategoryPrivate.value,
      project_id: project.value.id,
    })
    showCreateCategory.value = false
    newCategoryName.value = ''
    newCategoryPrivate.value = false
  } finally {
    creatingCategory.value = false
  }
}

function openCreateChannelInCategory(categoryId: string | null, type: 'text' | 'voice' = 'text') {
  newChannelCategoryId.value = categoryId
  newChannelType.value = type
  showCreateChannel.value = true
}

function selectChannel(channelId: string) {
  channelsStore.setCurrentChannel(channelId)
  router.push({ name: 'project-channel', params: { projectId: route.params.projectId, channelId } })
}

function isActive(channelId: string): boolean {
  return route.params.channelId === channelId
}

function goBack() {
  const wsId = workspaceStore.currentWorkspaceId
  projectsStore.currentProjectId = null
  if (wsId) channelsStore.fetchChannels(wsId)
  router.push({ name: 'channels' })
}

function goToTasks() {
  const pid = route.params.projectId as string
  router.push(`/projects/${pid}/tasks`)
}
</script>

<template>
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <!-- Header: название проекта + кнопка назад -->
    <div class="px-3 py-2.5 border-b border-subtle flex items-center gap-2 flex-shrink-0">
      <button
        class="p-1 rounded text-muted hover:text-primary hover:bg-elevated transition-colors"
        title="Назад в воркспейс"
        @click="goBack"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <span
        class="w-10 h-10 rounded flex-shrink-0 overflow-hidden flex items-center justify-center text-lg font-semibold"
        :style="!project?.icon_url ? { backgroundColor: 'var(--accent-600-hex)', color: 'var(--accent-300-hex)' } : {}"
      >
        <img v-if="project?.icon_url" :src="project.icon_url" :alt="project.name" class="w-full h-full object-cover" />
        <template v-else>{{ project?.name?.[0]?.toUpperCase() }}</template>
      </span>
      <span class="font-semibold text-primary truncate flex-1 text-sm">{{ project?.name ?? 'Проект' }}</span>
    </div>

    <!-- Список каналов -->
    <div class="flex-1 overflow-y-auto p-3 space-y-2">
      <div v-if="project?.is_archived" class="px-2 py-1 bg-warning/10 text-warning text-xs rounded">
        Проект архивирован
      </div>

      <!-- Текстовые каналы по категориям -->
      <template v-for="group in groupedTextChannels" :key="group.id ?? '__uncategorized_text'">
        <div v-if="group.channels.length > 0 || (canManage && group.id !== null)">
          <div class="px-2 mb-1 flex items-center justify-between group/header">
            <button
              class="flex items-center gap-1 text-xs font-semibold text-subtle uppercase tracking-wider hover:text-tertiary transition-colors min-w-0"
              @click="group.id ? toggleCategory(group.id) : undefined"
            >
              <svg
                v-if="group.id"
                class="w-3 h-3 shrink-0 transition-transform"
                :class="collapsedCategories[group.id] ? '-rotate-90' : ''"
                fill="currentColor" viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
              </svg>
              <span class="truncate">{{ group.name ?? 'Текстовые каналы' }}</span>
            </button>
            <button
              v-if="canManage"
              class="opacity-0 group-hover/header:opacity-100 w-4 h-4 flex items-center justify-center text-subtle hover:text-primary transition-all shrink-0"
              title="Добавить канал"
              @click.stop="openCreateChannelInCategory(group.id, 'text')"
            >
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" class="w-4 h-4">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
          <div v-if="!group.id || !collapsedCategories[group.id]" class="space-y-0.5">
            <button
              v-for="ch in group.channels"
              :key="ch.id"
              class="nav-item w-full"
              :class="[{ active: isActive(ch.id) }, group.id ? 'pl-4' : '']"
              @click="selectChannel(ch.id)"
            >
              <span class="relative w-5 h-5 flex-shrink-0 flex items-center justify-center">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                </svg>
                <svg v-if="(ch as any).is_private" class="absolute -bottom-0.5 -right-0.5 w-3 h-3 text-subtle bg-surface rounded-sm" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
                </svg>
              </span>
              <span class="truncate text-sm" :class="(ch as any).unread_count > 0 ? 'font-semibold' : ''">{{ ch.name }}</span>
              <span v-if="(ch as any).unread_count > 0" class="ml-auto px-1.5 py-0.5 text-xs font-semibold bg-accent-light text-white rounded-full min-w-[20px] text-center">
                {{ (ch as any).unread_count > 99 ? '99+' : (ch as any).unread_count }}
              </span>
            </button>
          </div>
        </div>
      </template>

      <!-- Голосовые каналы -->
      <template v-for="group in groupedVoiceChannels" :key="(group.id ?? '__uncategorized_voice') + '_voice'">
        <div v-if="group.channels.length > 0 || canManage">
          <div class="px-2 mb-1 flex items-center justify-between group/header">
            <button
              class="flex items-center gap-1 text-xs font-semibold text-subtle uppercase tracking-wider hover:text-tertiary transition-colors min-w-0"
              @click="group.id ? toggleCategory(group.id + '_v') : undefined"
            >
              <svg v-if="group.id" class="w-3 h-3 shrink-0 transition-transform" :class="collapsedCategories[group.id + '_v'] ? '-rotate-90' : ''" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
              </svg>
              <span class="truncate">{{ group.name ?? 'Голосовые каналы' }}</span>
            </button>
            <button
              v-if="canManage"
              class="opacity-0 group-hover/header:opacity-100 w-4 h-4 flex items-center justify-center text-subtle hover:text-primary transition-all shrink-0"
              title="Добавить голосовой канал"
              @click.stop="openCreateChannelInCategory(group.id, 'voice')"
            >
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" class="w-4 h-4">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
          <div v-if="!group.id || !collapsedCategories[group.id + '_v']" class="space-y-0.5">
            <button
              v-for="ch in group.channels"
              :key="ch.id"
              class="w-full px-2 py-1.5 flex items-center gap-2 rounded-lg transition-colors text-left text-muted hover:text-primary hover:bg-elevated"
              :class="group.id ? 'pl-4' : ''"
              @click="selectChannel(ch.id)"
            >
              <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
              </svg>
              <span class="truncate text-sm">{{ ch.name }}</span>
            </button>
          </div>
        </div>
      </template>

      <div v-if="!hasAnyChannels && !channelsStore.loading" class="text-center py-8">
        <p class="text-subtle text-sm">Нет каналов</p>
        <button v-if="canManage" class="text-xs text-accent mt-2 hover:underline" @click="openCreateChannelInCategory(null)">
          Создать первый канал
        </button>
      </div>
    </div>

    <!-- Нижняя навигация — как в воркспейсе -->
    <div class="p-3 border-t border-subtle space-y-1 flex-shrink-0">
      <!-- Задачи -->
      <button class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors" @click="goToTasks">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
        </svg>
        Задачи
      </button>

      <template v-if="canManage">
        <!-- Создать канал -->
        <button class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors" @click="openCreateChannelInCategory(null)">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Создать канал
        </button>

        <!-- Создать категорию -->
        <button class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors" @click="showCreateCategory = true">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7a2 2 0 012-2h4l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
          </svg>
          Создать категорию
        </button>

        <!-- Настройки проекта -->
        <button class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors" @click="showSettings = true">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          Настройки проекта
        </button>
      </template>
    </div>
  </div>

  <!-- Модал создания категории -->
  <Modal :open="showCreateCategory" title="Создать категорию" @close="showCreateCategory = false">
    <form class="space-y-4" @submit.prevent="createCategory">
      <Input v-model="newCategoryName" label="Название категории" placeholder="Разработка" required />
      <div
        class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
        :class="newCategoryPrivate ? 'bg-accent-dim border border-accent-dim' : 'bg-elevated border border-default'"
        @click="newCategoryPrivate = !newCategoryPrivate"
      >
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <div>
            <p class="text-sm font-medium text-primary">Приватная категория</p>
            <p class="text-xs text-subtle">Только приглашённые участники проекта</p>
          </div>
        </div>
        <div class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0" :class="newCategoryPrivate ? 'bg-accent' : 'bg-muted-fill'">
          <span class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform" :class="newCategoryPrivate ? 'translate-x-[18px]' : 'translate-x-[3px]'" />
        </div>
      </div>
      <div class="flex gap-3 pt-2">
        <Button variant="secondary" class="flex-1" @click="showCreateCategory = false">Отмена</Button>
        <Button type="submit" :loading="creatingCategory" class="flex-1">Создать</Button>
      </div>
    </form>
  </Modal>

  <!-- Модал создания канала -->
  <Modal :open="showCreateChannel" title="Создать канал в проекте" @close="showCreateChannel = false">
    <form class="space-y-4" @submit.prevent="createChannel">
      <Input v-model="newChannelName" label="Название канала" placeholder="general" @input="onChannelNameInput" />

      <div v-if="channelsStore.categories.length > 0">
        <label class="block text-sm font-medium text-tertiary mb-1.5">Категория</label>
        <select v-model="newChannelCategoryId" class="w-full px-3 py-2 bg-surface border border-default rounded-lg text-primary focus:border-accent focus:outline-none text-sm">
          <option :value="null">Без категории</option>
          <option v-for="cat in channelsStore.categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </select>
      </div>

      <div class="space-y-2">
        <label class="block text-sm font-medium text-tertiary">Тип канала</label>
        <div class="flex gap-3">
          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="newChannelType" type="radio" value="text" class="w-4 h-4">
            <span class="text-sm text-secondary">Текстовый</span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input v-model="newChannelType" type="radio" value="voice" class="w-4 h-4">
            <span class="text-sm text-secondary">Голосовой</span>
          </label>
        </div>
      </div>

      <div
        class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
        :class="newChannelPrivate ? 'bg-accent-dim border border-accent-dim' : 'bg-elevated border border-default'"
        @click="newChannelPrivate = !newChannelPrivate"
      >
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
          <div>
            <p class="text-sm font-medium text-primary">Приватный канал</p>
            <p class="text-xs text-subtle">Только приглашённые участники проекта</p>
          </div>
        </div>
        <div class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0" :class="newChannelPrivate ? 'bg-accent' : 'bg-muted-fill'">
          <span class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform" :class="newChannelPrivate ? 'translate-x-[18px]' : 'translate-x-[3px]'" />
        </div>
      </div>

      <div class="flex gap-3 pt-2">
        <Button variant="secondary" class="flex-1" @click="showCreateChannel = false">Отмена</Button>
        <Button type="submit" :loading="creatingChannel" class="flex-1">Создать</Button>
      </div>
    </form>
  </Modal>

  <ProjectSettingsModal
    v-if="showSettings && project"
    :project="project"
    @close="showSettings = false"
  />
</template>
