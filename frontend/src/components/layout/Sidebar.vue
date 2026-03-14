<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore, useWorkspaceStore, useChannelsStore } from '@/stores'
import { useProjectsStore } from '@/stores/projects'
import { useNavigationStore } from '@/stores/navigation'
import { Avatar, Modal, Input, Button } from '@/components/ui'
import ChannelList from './ChannelList.vue'
import CallPanel from '@/components/calls/CallPanel.vue'
import WorkspaceSettingsModal from '@/components/workspace/WorkspaceSettingsModal.vue'
import UserSettingsModal from '@/components/workspace/UserSettingsModal.vue'
import ProjectSidebar from '@/components/project/ProjectSidebar.vue'
import { authApi } from '@/api/auth'
import type { UserStatusValue } from '@/api/auth'
import { useDMStore } from '@/stores/dm'
import { RouterLink } from 'vue-router'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const projectsStore = useProjectsStore()
const navigationStore = useNavigationStore()
const dmStoreInst = useDMStore()

const dmListSidebar = computed(() => dmStoreInst.dmList)

const STATUS_COLORS: Record<string, string> = {
  online: '#3fb950',
  away: '#d29922',
  dnd: '#f85149',
  offline: '#6e7681',
}

function statusColor(status: string): string {
  return STATUS_COLORS[status] ?? STATUS_COLORS.offline
}

const isProjectRoute = computed(() =>
  navigationStore.activeSection === 'project' ||
  route.name === 'project' || route.name === 'project-channel' || route.name === 'project-tasks'
)
const isDMRoute = computed(() =>
  navigationStore.activeSection === 'dm' ||
  route.name === 'dm' || route.name === 'dm-channel'
)

const showCreateChannel = ref(false)
const showWorkspaceSettings = ref(false)
const showUserSettings = ref(false)
const showCreateProject = ref(false)
const newProjectName = ref('')
const creatingProject = ref(false)

const currentUserRole = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  const members = workspaceStore.membersMap[wsId] ?? []
  return members.find(m => m.user_id === authStore.user?.id)?.role ?? null
})

const isAdmin = computed(() =>
  currentUserRole.value === 'owner' || currentUserRole.value === 'admin'
)

const canCreateProject = computed(() => {
  const role = currentUserRole.value
  if (role === 'owner' || role === 'admin') return true
  // TODO: проверка create_projects через effectivePermissions
  return false
})

async function createProject() {
  if (!newProjectName.value.trim() || !workspaceStore.currentWorkspaceId) return
  creatingProject.value = true
  try {
    const project = await projectsStore.createProject(workspaceStore.currentWorkspaceId, {
      name: newProjectName.value.trim(),
    })
    showCreateProject.value = false
    newProjectName.value = ''
    router.push({ name: 'project', params: { projectId: project.id } })
  } finally {
    creatingProject.value = false
  }
}
const newChannelName = ref('')
const newChannelType = ref<'text' | 'voice'>('text')
const newChannelPrivate = ref(false)
const newChannelCategoryId = ref<string | null>(null)
const creatingChannel = ref(false)

const showCreateCategory = ref(false)
const newCategoryName = ref('')
const newCategoryPrivate = ref(false)
const creatingCategory = ref(false)

async function createCategory() {
  if (!newCategoryName.value.trim() || !workspaceStore.currentWorkspaceId) return
  creatingCategory.value = true
  try {
    await channelsStore.createCategory(workspaceStore.currentWorkspaceId, {
      name: newCategoryName.value.trim(),
      is_private: newCategoryPrivate.value,
    })
    showCreateCategory.value = false
    newCategoryName.value = ''
    newCategoryPrivate.value = false
  } finally {
    creatingCategory.value = false
  }
}

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

// Status context menu
const showStatusMenu = ref(false)
const statusMenuAnchorRef = ref<HTMLElement | null>(null)

const STATUS_OPTIONS: { value: UserStatusValue; label: string; color: string }[] = [
  { value: 'online',  label: 'В сети',         color: '#3fb950' },
  { value: 'away',    label: 'Отошёл',          color: '#d29922' },
  { value: 'dnd',     label: 'Не беспокоить',   color: '#f85149' },
  { value: 'offline', label: 'Невидимка',       color: '#6e7681' },
]

const currentStatusColor = computed(() => {
  const s = authStore.user?.status ?? 'offline'
  return STATUS_OPTIONS.find(o => o.value === s)?.color ?? '#6e7681'
})

async function setStatus(status: UserStatusValue) {
  showStatusMenu.value = false
  try {
    const updated = await authApi.updateStatus(status, authStore.user?.custom_status ?? null)
    authStore.user = updated
  } catch { /* тихо */ }
}

function closeStatusMenu(e: MouseEvent) {
  if (statusMenuAnchorRef.value && !statusMenuAnchorRef.value.contains(e.target as Node)) {
    showStatusMenu.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', closeStatusMenu))
onBeforeUnmount(() => document.removeEventListener('mousedown', closeStatusMenu))

async function logout() {
  await authStore.logout()
  await router.push('/login')
}
</script>

<template>
  <aside class="w-60 flex flex-col bg-surface border-r border-subtle">
    <!-- Режим проекта -->
    <template v-if="isProjectRoute">
      <ProjectSidebar class="flex-1 min-h-0" />
    </template>

    <!-- DM: список диалогов — пустой placeholder, реальный список в DMView -->
    <template v-else-if="isDMRoute">
      <div class="flex-1 flex flex-col">
        <div class="px-3 py-3 border-b border-subtle">
          <p class="text-xs font-semibold text-subtle uppercase tracking-wider">Личные сообщения</p>
        </div>
        <RouterLink
          v-for="dm in dmListSidebar"
          :key="dm.channelId"
          :to="{ name: 'dm-channel', params: { channelId: dm.channelId } }"
          class="flex items-center gap-2.5 px-3 py-2 hover:bg-elevated transition-colors"
          active-class="bg-elevated"
        >
          <div class="relative shrink-0">
            <Avatar :name="dm.peer.displayName" :src="dm.peer.avatarUrl ?? undefined" size="sm" />
            <span
              class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-[var(--bg-surface)]"
              :style="{ background: statusColor(dm.peer.status) }"
            />
          </div>
          <span
            class="text-sm truncate flex-1"
            :class="dm.unreadCount > 0 ? 'text-primary font-semibold' : 'text-secondary'"
          >{{ dm.peer.displayName }}</span>
          <span
            v-if="dm.unreadCount > 0"
            class="ml-auto shrink-0 min-w-[18px] h-[18px] px-1 rounded-full bg-accent text-white text-[10px] font-bold flex items-center justify-center"
          >{{ dm.unreadCount > 99 ? '99+' : dm.unreadCount }}</span>
        </RouterLink>
      </div>
    </template>

    <!-- Обычный режим воркспейса -->
    <template v-else>
      <div class="flex-1 overflow-y-auto">
        <ChannelList @create-channel="showCreateChannel = true" />

        <!-- Секция проектов -->
        <div class="mt-2">
          <ProjectList />
          <div v-if="canCreateProject" class="px-3 mt-1">
            <button
              class="w-full text-left px-2 py-1 text-xs text-muted hover:text-primary transition-colors"
              @click="showCreateProject = true"
            >
              + Создать проект
            </button>
          </div>
        </div>
      </div>

      <CallPanel />

      <div class="p-3 border-t border-subtle space-y-1">
        <RouterLink
          to="/tasks"
          class="nav-item"
          active-class="active"
        >
          <svg
            class="w-4 h-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
            />
          </svg>
          Задачи
        </RouterLink>
        <template v-if="isAdmin">
          <button
            class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors"
            @click="showCreateChannel = true"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            Создать канал
          </button>
          <button
            class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors"
            @click="showCreateCategory = true"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M3 7a2 2 0 012-2h4l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z"
              />
            </svg>
            Создать категорию
          </button>
          <button
            class="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-primary hover:bg-elevated transition-colors"
            @click="showWorkspaceSettings = true"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
            Настройки воркспейса
          </button>
        </template>
      </div>
    </template>

    <!-- User panel — всегда виден -->
    <div class="relative p-3 border-t border-subtle">
      <!-- Status context menu -->
      <Transition
        enter-active-class="transition-all duration-150 origin-bottom-left"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition-all duration-100 origin-bottom-left"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="showStatusMenu"
          class="absolute bottom-full left-3 mb-1 w-52 rounded-xl bg-overlay border border-default shadow-lg overflow-hidden z-50"
        >
          <div class="px-3 py-2 border-b border-default">
            <p class="text-xs font-semibold text-subtle uppercase tracking-wider">Статус</p>
          </div>
          <div class="py-1">
            <button
              v-for="opt in STATUS_OPTIONS"
              :key="opt.value"
              type="button"
              class="flex items-center gap-3 w-full px-3 py-2 text-sm text-secondary hover:bg-elevated transition-colors"
              :class="authStore.user?.status === opt.value ? 'text-primary font-medium' : ''"
              @click="setStatus(opt.value)"
            >
              <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ background: opt.color }" />
              {{ opt.label }}
              <svg
                v-if="authStore.user?.status === opt.value"
                class="ml-auto w-3.5 h-3.5 text-accent"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            </button>
          </div>
        </div>
      </Transition>

      <div ref="statusMenuAnchorRef" class="flex items-center gap-2">
        <!-- Avatar + name block — opens status menu on click -->
        <button
          type="button"
          class="flex items-center gap-2 flex-1 min-w-0 rounded-lg px-1 py-1 hover:bg-elevated transition-colors text-left"
          @click="showStatusMenu = !showStatusMenu"
        >
          <div class="relative shrink-0">
            <Avatar
              v-if="authStore.user"
              :name="authStore.user.display_name"
              :src="authStore.user.avatar_url"
              size="sm"
            />
            <!-- Status dot -->
            <span
              class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[var(--bg-surface)]"
              :style="{ background: currentStatusColor }"
            />
          </div>
          <div class="min-w-0">
            <p class="text-sm font-medium text-primary truncate leading-tight">
              {{ authStore.user?.display_name }}
            </p>
            <p class="text-xs text-subtle truncate leading-tight">
              {{ authStore.user?.custom_status || authStore.user?.email }}
            </p>
          </div>
        </button>

        <!-- Settings -->
        <button
          class="btn-ghost p-1.5 rounded-lg text-muted hover:text-primary shrink-0"
          title="Настройки"
          @click="showUserSettings = true"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>

        <!-- Logout -->
        <button
          class="btn-ghost p-1.5 rounded-lg text-muted hover:text-primary shrink-0"
          title="Выйти"
          @click="logout"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </div>

    <WorkspaceSettingsModal
      :open="showWorkspaceSettings"
      @close="showWorkspaceSettings = false"
    />

    <UserSettingsModal
      :open="showUserSettings"
      @close="showUserSettings = false"
    />

    <Modal
      :open="showCreateChannel"
      title="Создать канал"
      @close="showCreateChannel = false"
    >
      <form
        class="space-y-4"
        @submit.prevent="createChannel"
      >
        <Input
          v-model="newChannelName"
          label="Название канала"
          placeholder="general"
          @input="onChannelNameInput"
        />

        <div class="space-y-2">
          <label class="block text-sm font-medium text-tertiary">Тип канала</label>
          <div class="flex gap-3">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="newChannelType"
                type="radio"
                value="text"
                class="w-4 h-4 text-accent-600 bg-elevated border-strong"
              >
              <span class="text-sm text-secondary">Текстовый</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="newChannelType"
                type="radio"
                value="voice"
                class="w-4 h-4 text-accent-600 bg-elevated border-strong"
              >
              <span class="text-sm text-secondary">Голосовой</span>
            </label>
          </div>
        </div>

        <div v-if="channelsStore.categories.length > 0">
          <label class="block text-sm font-medium text-tertiary mb-1.5">Категория</label>
          <select
            v-model="newChannelCategoryId"
            class="w-full px-3 py-2 bg-surface border border-default rounded-lg text-primary focus:border-accent focus:outline-none text-sm"
          >
            <option :value="null">
              Без категории
            </option>
            <option
              v-for="cat in channelsStore.categories"
              :key="cat.id"
              :value="cat.id"
            >
              {{ cat.name }}
            </option>
          </select>
        </div>

        <div
          class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
          :class="newChannelPrivate ? 'bg-accent-dim border border-accent-dim' : 'bg-elevated border border-default'"
          @click="newChannelPrivate = !newChannelPrivate"
        >
          <div class="flex items-center gap-2">
            <svg
              class="w-4 h-4 text-muted shrink-0"
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
              <p class="text-xs text-subtle">
                Только приглашённые участники
              </p>
            </div>
          </div>
          <div
            class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0"
            :class="newChannelPrivate ? 'bg-accent' : 'bg-muted-fill'"
          >
            <span
              class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
              :class="newChannelPrivate ? 'translate-x-[18px]' : 'translate-x-[3px]'"
            />
          </div>
        </div>

        <div class="flex gap-3 pt-2">
          <Button
            variant="secondary"
            class="flex-1"
            @click="showCreateChannel = false"
          >
            Отмена
          </Button>
          <Button
            type="submit"
            :loading="creatingChannel"
            class="flex-1"
          >
            Создать
          </Button>
        </div>
      </form>
    </Modal>

    <!-- Модал создания проекта -->
    <Modal :open="showCreateProject" title="Создать проект" @close="showCreateProject = false">
      <form @submit.prevent="createProject">
        <div class="space-y-4">
          <div>
            <label class="block text-sm text-muted mb-1">Название проекта</label>
            <Input
              v-model="newProjectName"
              placeholder="Мой проект"
              maxlength="100"
              required
            />
          </div>
          <div class="flex gap-3 pt-2">
            <Button variant="secondary" class="flex-1" @click="showCreateProject = false">Отмена</Button>
            <Button type="submit" :loading="creatingProject" class="flex-1">Создать</Button>
          </div>
        </div>
        </form>
    </Modal>

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
              <p class="text-xs text-subtle">Только приглашённые участники</p>
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
  </aside>
</template>
