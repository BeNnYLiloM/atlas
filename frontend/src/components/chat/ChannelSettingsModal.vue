<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useChannelsStore } from '@/stores'
import { useWorkspaceStore } from '@/stores'
import { useAuthStore } from '@/stores'
import { useProjectsStore } from '@/stores/projects'
import { channelsApi } from '@/api'
import { Button, Input, Select, Checkbox } from '@/components/ui'
import type { ChannelPermissions, NotificationLevel, WorkspaceRole } from '@/types'

const props = defineProps<{
  open: boolean
  channelId: string
}>()

const emit = defineEmits<{
  close: []
}>()

const channelsStore = useChannelsStore()
const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()
const projectsStore = useProjectsStore()

type Tab = 'overview' | 'permissions' | 'notifications'
const activeTab = ref<Tab>('overview')

const channel = computed(() =>
  channelsStore.channels.find(c => c.id === props.channelId) ?? null
)

// --- Права доступа ---
const currentUserRole = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  const members = workspaceStore.membersMap[wsId] ?? []
  return members.find(m => m.user_id === authStore.user?.id)?.role ?? null
})

const canEditChannel = computed(() => {
  const r = currentUserRole.value
  return r === 'owner' || r === 'admin'
})

// --- Overview tab ---
const name = ref('')
const topic = ref('')
const isPrivate = ref(false)
const slowmode = ref(0)

const saving = ref(false)
const saveError = ref('')
const saveSuccess = ref(false)

const showDeleteConfirm = ref(false)
const deleting = ref(false)

async function saveOverview() {
  if (!channel.value) return
  saving.value = true
  saveError.value = ''
  saveSuccess.value = false
  try {
    await channelsStore.updateChannelSettings(channel.value.id, {
      name: name.value.trim() || undefined,
      topic: topic.value.trim() || null,
      slowmode_seconds: slowmode.value,
    })
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 2000)
  } catch {
    saveError.value = 'Не удалось сохранить изменения'
  } finally {
    saving.value = false
  }
}

async function deleteCurrentChannel() {
  if (!channel.value) return
  deleting.value = true
  try {
    await channelsStore.deleteChannel(channel.value.id)
    emit('close')
  } catch {
    saveError.value = 'Не удалось удалить канал'
  } finally {
    deleting.value = false
    showDeleteConfirm.value = false
  }
}

// --- Permissions tab ---
const permissions = ref<ChannelPermissions>({ roles: [], users: [] })

watch(
  () => [props.open, props.channelId, channel.value?.name],
  ([open]) => {
    if (open && channel.value) {
      name.value = channel.value.name
      topic.value = channel.value.topic ?? ''
      isPrivate.value = channel.value.is_private
      slowmode.value = channel.value.slowmode_seconds
      saveError.value = ''
      saveSuccess.value = false
      activeTab.value = 'overview'
      permissions.value = { roles: [], users: [] }
    }
  },
  { immediate: true },
)
const permsLoading = ref(false)
const permsError = ref('')
const permsSaving = ref(false)
const permsSaved = ref(false)

// Локальные черновые наборы (Set roleId / userId)
const draftRoleIds = ref<Set<string>>(new Set())
const draftUserIds = ref<Set<string>>(new Set())

const wsRoles = ref<WorkspaceRole[]>([])
const userSearch = ref('')

async function fetchPermissions() {
  permsLoading.value = true
  permsError.value = ''
  try {
    const data = await channelsApi.getPermissions(props.channelId)
    permissions.value = { roles: data?.roles ?? [], users: data?.users ?? [] }
    // Синхронизируем черновик с сервером
    draftRoleIds.value = new Set(permissions.value.roles.map(r => r.role_id))
    draftUserIds.value = new Set(permissions.value.users.map(u => u.user_id))
  } catch (err) {
    console.error('[fetchPermissions]', err)
    permsError.value = 'Не удалось загрузить права'
  } finally {
    permsLoading.value = false
  }
}

watch(activeTab, (tab) => {
  if (tab === 'permissions') {
    fetchPermissions()
    loadWsRoles()
    if (channel.value?.project_id && !projectsStore.membersMap[channel.value.project_id]?.length) {
      projectsStore.fetchMembers(channel.value.project_id)
    }
  }
})

async function loadWsRoles() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId || wsRoles.value.length) return
  const { rolesApi } = await import('@/api/roles')
  wsRoles.value = await rolesApi.list(wsId)
}

// Если канал принадлежит проекту — ограничиваем список только участниками проекта
const isProjectChannel = computed(() => !!channel.value?.project_id)

const projectMemberUserIds = computed<Set<string>>(() => {
  if (!isProjectChannel.value || !channel.value?.project_id) return new Set()
  const members = projectsStore.membersMap[channel.value.project_id] ?? []
  return new Set(members.map(m => m.user_id))
})

// Роли: для канала проекта показываем, но фильтруем по участникам проекта на бэкенде
const allSelectableRoles = computed(() => {
  return wsRoles.value.filter(r => !r.is_system)
})

// Количество участников проекта с данной ролью (для подсказки в UI)
function projectMembersWithRole(roleId: string): number {
  if (!isProjectChannel.value || !channel.value?.project_id) return 0
  const members = projectsStore.membersMap[channel.value.project_id] ?? []
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return 0
  const wsMembers = workspaceStore.membersMap[wsId] ?? []
  const memberIds = new Set(members.map(m => m.user_id))
  return wsMembers.filter(m => memberIds.has(m.user_id) && m.custom_roles?.some(r => r.id === roleId)).length
}

// Участники: для канала проекта — только участники этого проекта (кроме owner/admin, у них доступ автоматически)
const allSelectableUsers = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return []
  const wsMembers = workspaceStore.membersMap[wsId] ?? []
  const q = userSearch.value.toLowerCase()

  if (isProjectChannel.value) {
    // Только участники проекта, не являющиеся owner/admin
    return wsMembers.filter(m =>
      m.role !== 'owner' &&
      m.role !== 'admin' &&
      projectMemberUserIds.value.has(m.user_id) &&
      (q === '' || (m.nickname ?? m.display_name ?? '').toLowerCase().includes(q))
    )
  }

  return wsMembers.filter(m =>
    m.role !== 'owner' &&
    m.role !== 'admin' &&
    (q === '' || (m.nickname ?? m.display_name ?? '').toLowerCase().includes(q))
  )
})

function toggleRole(roleId: string) {
  if (draftRoleIds.value.has(roleId)) {
    draftRoleIds.value.delete(roleId)
  } else {
    draftRoleIds.value.add(roleId)
  }
  draftRoleIds.value = new Set(draftRoleIds.value)
}

function toggleUser(userId: string) {
  if (draftUserIds.value.has(userId)) {
    draftUserIds.value.delete(userId)
  } else {
    draftUserIds.value.add(userId)
  }
  draftUserIds.value = new Set(draftUserIds.value)
}

async function savePermissions() {
  permsSaving.value = true
  permsError.value = ''
  try {
    const prevRoleIds = new Set(permissions.value.roles.map(r => r.role_id))
    const prevUserIds = new Set(permissions.value.users.map(u => u.user_id))

    // Роли: добавляем новые, удаляем убранные
    const rolesToAdd = [...draftRoleIds.value].filter(id => !prevRoleIds.has(id))
    const rolesToRemove = [...prevRoleIds].filter(id => !draftRoleIds.value.has(id))
    await Promise.all([
      ...rolesToAdd.map(id => channelsApi.addRolePermission(props.channelId, id)),
      ...rolesToRemove.map(id => channelsApi.removeRolePermission(props.channelId, id)),
    ])

    // Участники: добавляем новых, удаляем убранных
    const usersToAdd = [...draftUserIds.value].filter(id => !prevUserIds.has(id))
    const usersToRemove = [...prevUserIds].filter(id => !draftUserIds.value.has(id))
    await Promise.all([
      ...usersToAdd.map(id => channelsApi.addUserPermission(props.channelId, id)),
      ...usersToRemove.map(id => channelsApi.removeUserPermission(props.channelId, id)),
    ])

    await fetchPermissions()
    permsSaved.value = true
    setTimeout(() => { permsSaved.value = false }, 2000)
  } catch {
    permsError.value = 'Не удалось сохранить изменения'
  } finally {
    permsSaving.value = false
  }
}

// Переключатель приватности из вкладки Permissions
async function saveOverviewPrivacyToggle() {
  if (!channel.value) return
  permsError.value = ''
  try {
    await channelsStore.updateChannelSettings(channel.value.id, { is_private: !channel.value.is_private })
    // Даём стору обновиться, затем подгружаем права
    await fetchPermissions()
  } catch {
    permsError.value = 'Не удалось изменить тип канала'
  }
}

function getInitials(n: string) {
  return n.split(' ').map(p => p[0]).join('').toUpperCase().slice(0, 2)
}

function normalizeChannelName(value: string) {
  return value.toLowerCase().replace(/\s+/g, '-').replace(/[^a-zа-яё0-9-]/gi, '')
}

function onChannelNameInput(e: Event) {
  const input = e.target as HTMLInputElement
  const normalized = normalizeChannelName(input.value)
  name.value = normalized
  // Принудительно ставим нормализованное значение в DOM
  input.value = normalized
}

const slowmodeOptions = [
  { value: 0, label: 'Выключен' },
  { value: 5, label: '5 секунд' },
  { value: 10, label: '10 секунд' },
  { value: 30, label: '30 секунд' },
  { value: 60, label: '1 минута' },
  { value: 300, label: '5 минут' },
  { value: 900, label: '15 минут' },
  { value: 3600, label: '1 час' },
]

// --- Notifications tab ---
const notifLevel = computed<NotificationLevel>(() =>
  channelsStore.getNotificationLevel(props.channelId)
)
const savingNotif = ref(false)
const notifError = ref('')

async function setNotifLevel(level: NotificationLevel) {
  if (level === notifLevel.value) return
  savingNotif.value = true
  notifError.value = ''
  try {
    await channelsStore.updateNotifications(props.channelId, level)
  } catch {
    notifError.value = 'Не удалось обновить уведомления'
  } finally {
    savingNotif.value = false
  }
}

const notifOptions: { value: NotificationLevel; label: string; description: string }[] = [
  { value: 'all', label: 'Все сообщения', description: 'Уведомлять о каждом новом сообщении' },
  { value: 'mentions', label: 'Только упоминания', description: 'Уведомлять только когда упоминают вас' },
  { value: 'nothing', label: 'Ничего', description: 'Не получать уведомлений из этого канала' },
]
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.open && channel"
        class="fixed inset-0 z-50 flex bg-base/80 backdrop-blur-sm"
        @click.self="emit('close')"
      >
        <div class="m-auto flex w-full max-w-2xl h-[520px] card overflow-hidden p-0">
          <!-- Sidebar -->
          <div class="w-48 bg-surface border-r border-default flex flex-col p-3 shrink-0">
            <p class="text-xs font-semibold text-muted uppercase px-2 mb-2 truncate">
              # {{ channel.name }}
            </p>

            <button
              v-for="tab in [
                { id: 'overview', label: 'Основные' },
                ...(canEditChannel ? [{ id: 'permissions', label: 'Права доступа' }] : []),
                { id: 'notifications', label: 'Уведомления' },
              ]"
              :key="tab.id"
              class="w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors"
              :class="activeTab === tab.id
                ? 'bg-overlay text-primary'
                : 'text-tertiary hover:bg-elevated hover:text-primary'"
              @click="activeTab = tab.id as Tab"
            >
              {{ tab.label }}
            </button>

            <div class="mt-auto pt-3 border-t border-default">
              <button
                class="w-full text-left px-2 py-1.5 rounded-md text-sm text-muted hover:bg-elevated transition-colors"
                @click="emit('close')"
              >
                Закрыть
              </button>
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto p-6">
            <!-- Overview -->
            <div
              v-if="activeTab === 'overview'"
              class="space-y-5"
            >
              <h2 class="text-lg font-semibold text-primary">
                Основные настройки
              </h2>

              <template v-if="canEditChannel">
                <div>
                  <label class="block text-sm font-medium text-tertiary mb-1.5">Название канала</label>
                  <Input
                    v-model="name"
                    placeholder="general"
                    maxlength="100"
                    @input="onChannelNameInput"
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-tertiary mb-1.5">Топик / Описание</label>
                  <textarea
                    v-model="topic"
                    rows="3"
                    maxlength="1024"
                    placeholder="О чём этот канал..."
                    class="w-full px-3 py-2 bg-surface border border-default rounded-lg text-primary placeholder-subtle focus:border-accent focus:outline-none resize-none text-sm"
                  />
                  <p class="text-xs text-subtle mt-1 text-right">
                    {{ topic.length }}/1024
                  </p>
                </div>

                <div>
                  <label class="block text-sm font-medium text-tertiary mb-1.5">Slow Mode</label>
                  <p class="text-xs text-subtle mb-2">
                    Ограничение на частоту отправки сообщений участниками
                  </p>
                  <Select
                    v-model="slowmode"
                    :options="slowmodeOptions"
                  />
                </div>

                <div
                  v-if="saveError"
                  class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
                >
                  <p class="text-sm text-red-400">
                    {{ saveError }}
                  </p>
                </div>
                <div
                  v-if="saveSuccess"
                  class="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg"
                >
                  <p class="text-sm text-emerald-400">
                    Изменения сохранены
                  </p>
                </div>

                <Button
                  :loading="saving"
                  @click="saveOverview"
                >
                  Сохранить
                </Button>

                <!-- Удаление канала -->
                <div class="border-t border-default pt-5 mt-2">
                  <h3 class="text-sm font-medium text-red-400 mb-3">
                    Опасная зона
                  </h3>
                  <div
                    v-if="!showDeleteConfirm"
                    class="flex items-center justify-between p-3 border border-red-500/20 rounded-lg bg-red-500/5"
                  >
                    <div>
                      <p class="text-sm font-medium text-primary">
                        Удалить канал
                      </p>
                      <p class="text-xs text-muted mt-0.5">
                        Это действие необратимо, все сообщения будут удалены
                      </p>
                    </div>
                    <button
                      class="px-3 py-1.5 text-sm font-medium text-red-400 border border-red-500/40 rounded-lg hover:bg-red-500/10 transition-colors"
                      @click="showDeleteConfirm = true"
                    >
                      Удалить
                    </button>
                  </div>
                  <div
                    v-else
                    class="p-3 border border-red-500/40 rounded-lg bg-red-500/10 space-y-3"
                  >
                    <p class="text-sm text-primary">
                      Вы уверены? Канал <span class="font-semibold">#{{ channel.name }}</span> и все его сообщения будут удалены навсегда.
                    </p>
                    <div class="flex gap-2">
                      <button
                        class="flex-1 px-3 py-1.5 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors disabled:opacity-50"
                        :disabled="deleting"
                        @click="deleteCurrentChannel"
                      >
                        {{ deleting ? 'Удаление...' : 'Да, удалить' }}
                      </button>
                      <button
                        class="px-3 py-1.5 text-sm font-medium text-tertiary border border-default rounded-lg hover:bg-overlay transition-colors"
                        @click="showDeleteConfirm = false"
                      >
                        Отмена
                      </button>
                    </div>
                  </div>
                </div>
              </template>

              <!-- Read-only для member -->
              <template v-else>
                <div class="space-y-3">
                  <div class="p-3 bg-surface rounded-lg">
                    <p class="text-xs text-muted mb-1">
                      Название
                    </p>
                    <p class="text-sm text-primary">
                      # {{ channel.name }}
                    </p>
                  </div>
                  <div
                    v-if="channel.topic"
                    class="p-3 bg-surface rounded-lg"
                  >
                    <p class="text-xs text-muted mb-1">
                      Топик
                    </p>
                    <p class="text-sm text-primary">
                      {{ channel.topic }}
                    </p>
                  </div>
                  <div class="p-3 bg-surface rounded-lg">
                    <p class="text-xs text-muted mb-1">
                      Тип
                    </p>
                    <p class="text-sm text-primary">
                      {{ channel.is_private ? 'Приватный' : 'Публичный' }}
                    </p>
                  </div>
                </div>
              </template>
            </div>

            <!-- Права доступа -->
            <div
              v-else-if="activeTab === 'permissions'"
              class="space-y-5"
            >
              <div>
                <h2 class="text-lg font-semibold text-primary">
                  Права канала
                </h2>
                <p class="text-xs text-muted mt-1">
                  Управляйте кто имеет доступ к этому каналу
                </p>
              </div>

              <!-- Приватный переключатель -->
              <div class="p-4 border border-default rounded-lg">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3">
                    <svg
                      class="w-5 h-5 text-muted shrink-0"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                      />
                    </svg>
                    <div>
                      <p class="text-sm font-medium text-primary">
                        Приватный канал
                      </p>
                      <p class="text-xs text-muted mt-0.5">
                        Только выбранные участники и роли могут просматривать его
                      </p>
                    </div>
                  </div>
                  <button
                    v-if="canEditChannel"
                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors shrink-0"
                    :class="channel.is_private ? 'bg-accent' : 'bg-overlay'"
                    @click="saveOverviewPrivacyToggle"
                  >
                    <span
                      class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform"
                      :class="channel.is_private ? 'translate-x-6' : 'translate-x-1'"
                    />
                  </button>
                  <div
                    v-else
                    class="text-xs px-2 py-1 rounded"
                    :class="channel.is_private ? 'bg-accent-dim-md text-accent-strong' : 'bg-elevated text-muted'"
                  >
                    {{ channel.is_private ? 'Приватный' : 'Публичный' }}
                  </div>
                </div>
              </div>

              <!-- Публичный канал: инфо-блок -->
              <div
                v-if="!channel.is_private"
                class="flex items-center gap-3 p-4 bg-surface rounded-lg border border-default"
              >
                <svg
                  class="w-5 h-5 text-accent shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                <p class="text-sm text-tertiary">
                  Права синхронизированы с воркспейсом — все участники имеют доступ к этому каналу
                </p>
              </div>

              <!-- Подсказка для канала проекта -->
              <div
                v-if="isProjectChannel"
                class="flex items-center gap-2 p-3 bg-surface rounded-lg border border-subtle text-xs text-muted"
              >
                <svg class="w-4 h-4 shrink-0 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Канал принадлежит проекту — в списке только участники этого проекта
              </div>

              <!-- Приватный: чекбокс-список -->
              <template v-if="channel.is_private">
                <div
                  v-if="permsError"
                  class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
                >
                  <p class="text-sm text-red-400">
                    {{ permsError }}
                  </p>
                </div>

                <div
                  v-if="permsLoading"
                  class="text-sm text-subtle text-center py-4"
                >
                  Загрузка...
                </div>
                <template v-else>
                  <!-- Роли -->
                  <div v-if="allSelectableRoles.length">
                    <div class="flex items-center justify-between mb-2">
                      <p class="text-xs font-semibold text-muted uppercase">
                        Роли
                      </p>
                      <span
                        v-if="isProjectChannel"
                        class="text-xs text-muted"
                        title="Роль предоставляет доступ только участникам проекта"
                      >только участники проекта</span>
                    </div>
                    <div class="space-y-1">
                      <div
                        v-for="role in allSelectableRoles"
                        :key="role.id"
                        class="flex items-center gap-3 px-3 py-2 rounded-lg bg-surface border border-subtle transition-colors"
                        :class="[
                          canEditChannel ? 'cursor-pointer hover:border-strong' : 'opacity-50 cursor-default',
                          draftRoleIds.has(role.id) ? 'border-accent-dim' : 'border-subtle',
                        ]"
                        @click="canEditChannel && toggleRole(role.id)"
                      >
                        <Checkbox
                          :model-value="draftRoleIds.has(role.id)"
                          :disabled="!canEditChannel"
                        />
                        <span
                          class="w-3 h-3 rounded-full shrink-0"
                          :style="{ backgroundColor: role.color }"
                        />
                        <span class="flex-1 text-sm text-primary">{{ role.name }}</span>
                        <span
                          v-if="isProjectChannel"
                          class="text-xs text-muted"
                          :title="`Участников проекта с этой ролью: ${projectMembersWithRole(role.id)}`"
                        >{{ projectMembersWithRole(role.id) }} участн.</span>
                      </div>
                    </div>
                  </div>

                  <!-- Участники -->
                  <div v-if="allSelectableUsers.length || userSearch">
                    <div class="flex items-center justify-between mb-2 mt-1">
                      <p class="text-xs font-semibold text-muted uppercase">
                        Участники
                      </p>
                      <input
                        v-model="userSearch"
                        type="text"
                        placeholder="Поиск..."
                        class="px-2 py-1 bg-surface border border-default rounded text-xs text-primary placeholder-subtle focus:border-accent focus:outline-none w-36"
                      >
                    </div>
                    <div class="space-y-1 max-h-52 overflow-y-auto">
                      <div
                        v-for="m in allSelectableUsers"
                        :key="m.user_id"
                        class="flex items-center gap-3 px-3 py-2 rounded-lg bg-surface border transition-colors"
                        :class="[
                          canEditChannel ? 'cursor-pointer hover:border-strong' : 'opacity-50 cursor-default',
                          draftUserIds.has(m.user_id) ? 'border-accent-dim' : 'border-subtle',
                        ]"
                        @click="canEditChannel && toggleUser(m.user_id)"
                      >
                        <Checkbox
                          :model-value="draftUserIds.has(m.user_id)"
                          :disabled="!canEditChannel"
                        />
                        <div class="w-6 h-6 rounded-full bg-accent flex items-center justify-center text-white text-xs font-semibold shrink-0">
                          {{ getInitials(m.nickname ?? m.display_name ?? '') }}
                        </div>
                        <span class="flex-1 text-sm text-primary truncate">{{ m.nickname ?? m.display_name }}</span>
                      </div>
                      <p
                        v-if="!allSelectableUsers.length"
                        class="text-xs text-subtle px-2 py-1"
                      >
                        Не найдено
                      </p>
                    </div>
                  </div>

                  <!-- Кнопка Сохранить -->
                  <div
                    v-if="canEditChannel"
                    class="pt-2 space-y-2"
                  >
                    <div
                      v-if="permsSaved"
                      class="p-2 bg-emerald-500/10 border border-emerald-500/20 rounded-lg"
                    >
                      <p class="text-xs text-emerald-400 text-center">
                        Изменения сохранены
                      </p>
                    </div>
                    <Button
                      :loading="permsSaving"
                      @click="savePermissions"
                    >
                      Сохранить права доступа
                    </Button>
                  </div>
                </template>
              </template>
            </div>

            <!-- Notifications -->
            <div
              v-else-if="activeTab === 'notifications'"
              class="space-y-4"
            >
              <h2 class="text-lg font-semibold text-primary">
                Уведомления
              </h2>
              <p class="text-sm text-muted">
                Настройте уведомления для канала <span class="text-primary">#{{ channel.name }}</span>
              </p>

              <div
                v-if="notifError"
                class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
              >
                <p class="text-sm text-red-400">
                  {{ notifError }}
                </p>
              </div>

              <div class="space-y-2">
                <button
                  v-for="opt in notifOptions"
                  :key="opt.value"
                  class="w-full flex items-center gap-3 p-4 rounded-lg border transition-colors text-left"
                  :class="notifLevel === opt.value
                    ? 'border-accent bg-accent-light/10'
                    : 'border-default bg-surface hover:border-strong'"
                  :disabled="savingNotif"
                  @click="setNotifLevel(opt.value)"
                >
                  <!-- Radio circle -->
                  <div
                    class="w-4 h-4 rounded-full border-2 flex items-center justify-center shrink-0"
                    :class="notifLevel === opt.value ? 'border-accent' : 'border-strong'"
                  >
                    <div
                      v-if="notifLevel === opt.value"
                      class="w-2 h-2 rounded-full bg-accent-light"
                    />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-primary">
                      {{ opt.label }}
                    </p>
                    <p class="text-xs text-muted mt-0.5">
                      {{ opt.description }}
                    </p>
                  </div>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

