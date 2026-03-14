<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useProjectsStore, useWorkspaceStore, useAuthStore, useChannelsStore } from '@/stores'
import AddProjectMemberModal from './AddProjectMemberModal.vue'
import AvatarCropper from '@/components/ui/AvatarCropper.vue'
import { projectsApi } from '@/api/projects'
import type { Project } from '@/types'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const projectsStore = useProjectsStore()
const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()
const channelsStore = useChannelsStore()

type Tab = 'general' | 'categories' | 'members' | 'danger'
const activeTab = ref<Tab>('general')

// Управление категориями
const selectedCategoryId = ref<string | null>(null)
const editingCategoryName = ref('')
const categorySaving = ref(false)
const categoryCreating = ref(false)
const newCategoryName = ref('')
const newCategoryPrivate = ref(false)

const selectedCategory = computed(() =>
  channelsStore.categories.find(c => c.id === selectedCategoryId.value) ?? null
)

function selectCategory(id: string, name: string) {
  selectedCategoryId.value = id
  editingCategoryName.value = name
}

async function saveCategoryName() {
  if (!selectedCategoryId.value || !editingCategoryName.value.trim()) return
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  categorySaving.value = true
  try {
    await channelsStore.renameCategory(wsId, selectedCategoryId.value, editingCategoryName.value.trim())
  } finally {
    categorySaving.value = false
  }
}

async function createCategory() {
  if (!newCategoryName.value.trim() || !workspaceStore.currentWorkspaceId) return
  categoryCreating.value = true
  try {
    const cat = await channelsStore.createCategory(workspaceStore.currentWorkspaceId, {
      name: newCategoryName.value.trim(),
      is_private: newCategoryPrivate.value,
      project_id: props.project.id,
    })
    newCategoryName.value = ''
    newCategoryPrivate.value = false
    selectCategory(cat.id, cat.name)
  } finally {
    categoryCreating.value = false
  }
}

async function deleteCategory() {
  if (!selectedCategoryId.value) return
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  await channelsStore.deleteCategory(wsId, selectedCategoryId.value)
  selectedCategoryId.value = null
  editingCategoryName.value = ''
}

async function togglePrivacy() {
  if (!selectedCategoryId.value || !selectedCategory.value) return
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  await channelsStore.toggleCategoryPrivacy(wsId, selectedCategoryId.value, !selectedCategory.value.is_private)
}

const name = ref(props.project.name)
const description = ref(props.project.description ?? '')
const saving = ref(false)
const showAddMember = ref(false)

const iconUrl = ref<string | null>(props.project.icon_url ?? null)
const uploadingIcon = ref(false)
const iconInputRef = ref<HTMLInputElement | null>(null)
const iconCropFile = ref<File | null>(null)

function onIconFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  iconCropFile.value = file
  if (iconInputRef.value) iconInputRef.value.value = ''
}

async function onIconCropConfirmed(blob: Blob) {
  iconCropFile.value = null
  uploadingIcon.value = true
  try {
    const croppedFile = new File([blob], 'icon.webp', { type: 'image/webp' })
    const updated = await projectsApi.uploadIcon(props.project.id, croppedFile)
    iconUrl.value = updated.icon_url ?? null
    projectsStore.onProjectUpdated(updated)
  } finally {
    uploadingIcon.value = false
  }
}

function onIconCropCancelled() {
  iconCropFile.value = null
}

async function removeIcon() {
  uploadingIcon.value = true
  try {
    const updated = await projectsStore.updateProject(props.project.id, { clear_icon_url: true })
    iconUrl.value = null
    projectsStore.onProjectUpdated(updated)
  } finally {
    uploadingIcon.value = false
  }
}

const currentWsMember = computed(() => {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return null
  return (workspaceStore.membersMap[wsId] ?? []).find(m => m.user_id === authStore.user?.id) ?? null
})

const canManageLeads = computed(() =>
  currentWsMember.value?.role === 'owner' ||
  currentWsMember.value?.role === 'admin'
)

onMounted(async () => {
  await projectsStore.fetchMembers(props.project.id)
})

const members = computed(() => projectsStore.membersMap[props.project.id] ?? [])

async function save() {
  saving.value = true
  try {
    await projectsStore.updateProject(props.project.id, {
      name: name.value,
      description: description.value || null,
    })
  } finally {
    saving.value = false
  }
}

async function toggleArchive() {
  if (props.project.is_archived) {
    await projectsStore.unarchiveProject(props.project.id)
  } else {
    await projectsStore.archiveProject(props.project.id)
  }
  emit('close')
}

async function deleteProject() {
  if (!confirm(`Удалить проект «${props.project.name}»? Это действие необратимо.`)) return
  await projectsStore.deleteProject(props.project.id)
  emit('close')
}

async function removeMember(userId: string) {
  await projectsStore.removeMember(props.project.id, userId)
}

async function toggleLead(userId: string, isLead: boolean) {
  await projectsStore.setLead(props.project.id, userId, !isLead)
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="emit('close')">
    <div class="bg-base rounded-xl shadow-xl w-full mx-4 flex flex-col"
      :class="activeTab === 'categories' ? 'max-w-2xl h-[520px]' : 'max-w-lg'"
    >
      <div class="flex items-center justify-between px-6 py-4 border-b border-subtle">
        <h2 class="font-semibold text-lg">Настройки проекта</h2>
        <button class="text-muted hover:text-primary" @click="emit('close')">✕</button>
      </div>

      <!-- Табы -->
      <div class="flex border-b border-subtle px-6">
        <button
          v-for="tab in ([['general','Основное'],['categories','Категории'],['members','Участники'],['danger','Опасная зона']] as [Tab, string][])"
          :key="tab[0]"
          class="py-3 px-2 mr-4 text-sm border-b-2 transition-colors"
          :class="activeTab === tab[0] ? 'border-accent text-primary' : 'border-transparent text-muted hover:text-primary'"
          @click="activeTab = tab[0]"
        >
          {{ tab[1] }}
        </button>
      </div>

      <div :class="activeTab === 'categories' ? 'flex-1 overflow-hidden flex' : 'px-6 py-4 max-h-96 overflow-y-auto'">
        <!-- Основное -->
        <template v-if="activeTab === 'general'">
          <div class="space-y-4">
            <!-- Иконка проекта -->
            <div>
              <label class="block text-sm text-muted mb-2">Иконка проекта</label>
              <div class="flex items-center gap-4">
                <div class="relative group">
                  <div
                    class="w-16 h-16 rounded-xl overflow-hidden flex items-center justify-center bg-elevated border border-subtle cursor-pointer"
                    :class="uploadingIcon ? 'opacity-50' : ''"
                    @click="iconInputRef?.click()"
                  >
                    <img v-if="iconUrl" :src="iconUrl" alt="Иконка проекта" class="w-full h-full object-cover" />
                    <span v-else class="text-2xl font-bold text-muted select-none">
                      {{ name[0]?.toUpperCase() ?? '?' }}
                    </span>
                    <div class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center rounded-xl">
                      <svg v-if="!uploadingIcon" class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                      <svg v-else class="w-5 h-5 text-white animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                      </svg>
                    </div>
                  </div>
                  <input
                    ref="iconInputRef"
                    type="file"
                    accept="image/*"
                    class="hidden"
                    @change="onIconFileChange"
                  />
                </div>
                <div class="text-xs text-subtle space-y-1">
                  <p>Нажмите на иконку для загрузки</p>
                  <p>PNG, JPG, GIF — до 5 МБ</p>
                  <button
                    v-if="iconUrl"
                    class="mt-1 text-red-400 hover:text-red-300 transition-colors"
                    :disabled="uploadingIcon"
                    @click="removeIcon"
                  >
                    Удалить иконку
                  </button>
                </div>
              </div>
            </div>

            <div>
              <label class="block text-sm text-muted mb-1">Название</label>
              <input v-model="name" type="text" maxlength="100"
                class="w-full bg-surface border border-subtle rounded px-3 py-2 text-sm focus:outline-none focus:border-accent" />
            </div>
            <div>
              <label class="block text-sm text-muted mb-1">Описание</label>
              <textarea v-model="description" rows="3" maxlength="500"
                class="w-full bg-surface border border-subtle rounded px-3 py-2 text-sm focus:outline-none focus:border-accent resize-none" />
            </div>
            <button
              :disabled="saving"
              class="px-4 py-2 bg-accent text-white rounded text-sm hover:bg-accent/90 disabled:opacity-50"
              @click="save"
            >
              {{ saving ? 'Сохранение...' : 'Сохранить' }}
            </button>
          </div>
        </template>

        <!-- Категории — двухпанельный layout как в настройках воркспейса -->
        <template v-if="activeTab === 'categories'">
          <!-- Левая панель: список -->
          <div class="w-52 border-r border-subtle flex flex-col bg-surface/50 shrink-0">
            <div class="p-3 border-b border-subtle">
              <span class="text-xs font-semibold text-muted uppercase">Категории</span>
            </div>
            <div class="flex-1 overflow-y-auto p-2 space-y-0.5">
              <button
                v-for="cat in channelsStore.categories"
                :key="cat.id"
                  class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left"
                    :class="selectedCategoryId === cat.id ? 'bg-overlay text-primary' : 'text-tertiary hover:bg-elevated'"
                    @click="selectCategory(cat.id, cat.name)"
                  >
                    <svg v-if="cat.is_private" class="w-3.5 h-3.5 shrink-0 text-muted" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
                    </svg>
                    <span class="truncate flex-1">{{ cat.name }}</span>
                  </button>
              <p v-if="channelsStore.categories.length === 0" class="text-xs text-subtle text-center py-3">
                Нет категорий
              </p>
            </div>
            <!-- Создать категорию -->
            <div class="p-3 border-t border-subtle space-y-2">
              <input
                v-model="newCategoryName"
                class="w-full px-2 py-1.5 bg-elevated border border-default rounded text-sm text-primary placeholder-subtle focus:border-accent focus:outline-none"
                placeholder="Новая категория..."
                @keyup.enter="createCategory"
              >
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-1.5 cursor-pointer text-xs text-muted">
                  <input v-model="newCategoryPrivate" type="checkbox" class="w-3.5 h-3.5 rounded">
                  Приватная
                </label>
                <button
                  class="px-3 py-1 text-sm bg-accent hover:bg-accent/90 text-white rounded-lg transition-colors disabled:opacity-50"
                  :disabled="categoryCreating || !newCategoryName.trim()"
                  @click="createCategory"
                >
                  {{ categoryCreating ? '...' : 'Создать' }}
                </button>
              </div>
            </div>
          </div>

          <!-- Правая панель: редактор выбранной категории -->
          <div class="flex-1 overflow-y-auto">
            <div v-if="!selectedCategory" class="flex items-center justify-center h-full text-subtle text-sm">
              Выберите категорию для настройки
            </div>
            <div v-else class="p-6 space-y-5">
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1 space-y-3">
                  <h3 class="text-base font-semibold text-primary">{{ selectedCategory.name }}</h3>
                  <div class="flex gap-2">
                    <input
                      v-model="editingCategoryName"
                      class="flex-1 px-3 py-1.5 bg-surface border border-default rounded-lg text-sm text-primary focus:border-accent focus:outline-none"
                      placeholder="Название категории"
                      @keyup.enter="saveCategoryName"
                    >
                    <button
                      class="px-3 py-1.5 text-sm bg-accent hover:bg-accent/90 text-white rounded-lg transition-colors disabled:opacity-50"
                      :disabled="categorySaving"
                      @click="saveCategoryName"
                    >
                      {{ categorySaving ? '...' : 'OK' }}
                    </button>
                  </div>
                </div>
                <button
                  class="p-1.5 text-muted hover:text-red-400 transition-colors mt-0.5 shrink-0"
                  title="Удалить категорию"
                  @click="deleteCategory"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>

              <!-- Приватность -->
              <div
                class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
                :class="selectedCategory.is_private ? 'bg-accent-dim border border-accent-dim' : 'bg-elevated border border-default'"
                @click="togglePrivacy"
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
                <div
                  class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0"
                  :class="selectedCategory.is_private ? 'bg-accent' : 'bg-muted-fill'"
                >
                  <span
                    class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
                    :class="selectedCategory.is_private ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                  />
                </div>
              </div>

              <!-- Каналы в категории -->
              <div>
                <p class="text-xs font-semibold text-muted uppercase mb-2">
                  Каналов в категории: {{ channelsStore.channels.filter(c => c.category_id === selectedCategoryId).length }}
                </p>
                <div class="space-y-0.5">
                  <div
                    v-for="ch in channelsStore.channels.filter(c => c.category_id === selectedCategoryId)"
                    :key="ch.id"
                    class="flex items-center gap-2 px-2 py-1.5 rounded-lg text-sm text-secondary"
                  >
                    <svg v-if="ch.type === 'text'" class="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                    </svg>
                    <svg v-else class="w-4 h-4 text-muted shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                    </svg>
                    <span class="truncate">{{ ch.name }}</span>
                  </div>
                  <p v-if="channelsStore.channels.filter(c => c.category_id === selectedCategoryId).length === 0" class="text-xs text-subtle px-2 py-1">
                    Нет каналов в этой категории
                  </p>
                </div>
              </div>
              <p class="text-xs text-subtle">При удалении категории каналы остаются, но теряют привязку к ней.</p>
            </div>
          </div>
        </template>

        <!-- Участники -->
        <template v-if="activeTab === 'members'">
          <div class="flex justify-end mb-3">
            <button
              class="px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent/90"
              @click="showAddMember = true"
            >
              + Добавить
            </button>
          </div>
          <div class="space-y-2">
            <div
              v-for="member in members"
              :key="member.user_id"
              class="flex items-center gap-3 py-2 border-b border-subtle last:border-0"
            >
              <div class="w-8 h-8 rounded-full bg-accent/20 flex items-center justify-center text-sm font-semibold flex-shrink-0">
                {{ member.display_name?.[0]?.toUpperCase() ?? '?' }}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm truncate">{{ member.display_name }}</p>
                <p v-if="member.is_lead" class="text-xs text-accent">Лид</p>
              </div>
              <div v-if="canManageLeads" class="flex gap-1">
                <button
                  class="text-xs px-2 py-1 rounded transition-colors"
                  :class="member.is_lead ? 'bg-accent/20 text-accent' : 'text-muted hover:text-accent'"
                  @click="toggleLead(member.user_id, member.is_lead)"
                >
                  {{ member.is_lead ? 'Снять лида' : 'Назначить лидом' }}
                </button>
                <button
                  class="text-xs px-2 py-1 rounded text-red-500 hover:bg-red-500/10"
                  @click="removeMember(member.user_id)"
                >
                  Убрать
                </button>
              </div>
            </div>
            <p v-if="members.length === 0" class="text-sm text-muted text-center py-4">Нет участников</p>
          </div>
        </template>

        <!-- Опасная зона -->
        <template v-if="activeTab === 'danger'">
          <div class="space-y-3">
            <div class="border border-subtle rounded-lg p-4">
              <p class="font-medium text-sm mb-1">
                {{ project.is_archived ? 'Разархивировать проект' : 'Архивировать проект' }}
              </p>
              <p class="text-xs text-muted mb-3">
                {{ project.is_archived
                  ? 'Участники снова смогут писать сообщения.'
                  : 'Проект станет доступен только для чтения.' }}
              </p>
              <button
                class="px-3 py-1.5 rounded text-sm border transition-colors"
                :class="project.is_archived
                  ? 'border-accent text-accent hover:bg-accent/10'
                  : 'border-yellow-500 text-yellow-500 hover:bg-yellow-500/10'"
                @click="toggleArchive"
              >
                {{ project.is_archived ? 'Разархивировать' : 'Архивировать' }}
              </button>
            </div>
            <div class="border border-red-500/30 rounded-lg p-4">
              <p class="font-medium text-sm mb-1 text-red-500">Удалить проект</p>
              <p class="text-xs text-muted mb-3">Удаление необратимо. Все каналы и сообщения будут потеряны.</p>
              <button
                class="px-3 py-1.5 rounded text-sm border border-red-500 text-red-500 hover:bg-red-500/10"
                @click="deleteProject"
              >
                Удалить
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>

  <AvatarCropper
    v-if="iconCropFile"
    :file="iconCropFile"
    shape="square"
    :output-size="512"
    @crop="onIconCropConfirmed"
    @cancel="onIconCropCancelled"
  />

  <AddProjectMemberModal
    v-if="showAddMember"
    :project-id="project.id"
    :workspace-id="project.workspace_id"
    :existing-member-ids="members.map((m) => m.user_id)"
    @close="showAddMember = false"
  />
</template>
