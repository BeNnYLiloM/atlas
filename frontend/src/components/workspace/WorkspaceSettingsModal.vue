<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useWorkspaceStore } from '@/stores'
import { useAuthStore } from '@/stores'
import { useChannelsStore } from '@/stores'
import { rolesApi, workspacesApi, categoriesApi } from '@/api'
import { Button, Input, Select } from '@/components/ui'
import type { WorkspaceMember, WorkspaceRole, RolePermissions, ChannelCategory, CategoryPermissions } from '@/types'
import { defaultPermissions } from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const workspaceStore = useWorkspaceStore()
const channelsStore = useChannelsStore()
const authStore = useAuthStore()

type Tab = 'general' | 'roles' | 'members' | 'categories'
const activeTab = ref<Tab>('general')

// --- General tab ---
const name = ref('')
const description = ref('')
const saving = ref(false)
const saveError = ref('')
const saveSuccess = ref(false)
const iconUploading = ref(false)
const iconInputRef = ref<HTMLInputElement | null>(null)

async function onIconChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !workspaceStore.currentWorkspaceId) return
  iconUploading.value = true
  try {
    const updated = await workspacesApi.uploadIcon(workspaceStore.currentWorkspaceId, file)
    workspaceStore.applyWorkspaceUpdate(updated)
  } catch {
    // ignore silently – user sees no change
  } finally {
    iconUploading.value = false
    if (iconInputRef.value) iconInputRef.value.value = ''
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      const ws = workspaceStore.currentWorkspace
      name.value = ws?.name ?? ''
      description.value = ws?.description ?? ''
      saveError.value = ''
      saveSuccess.value = false
      activeTab.value = 'general'
      roles.value = []
      selectedRole.value = null
    }
  },
)

async function saveGeneral() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return

  saving.value = true
  saveError.value = ''
  saveSuccess.value = false

  try {
    await workspaceStore.updateWorkspace(wsId, {
      name: name.value.trim() || undefined,
      description: description.value.trim() || null,
    })
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 2000)
  } catch {
    saveError.value = 'Не удалось сохранить изменения'
  } finally {
    saving.value = false
  }
}

async function deleteWorkspace() {
  if (!workspaceStore.currentWorkspaceId) return
  if (!confirm(`Удалить воркспейс «${workspaceStore.currentWorkspace?.name}»? Это действие необратимо.`)) return
}

// --- Categories tab ---
const newCategoryName = ref('')
const newCategoryPrivate = ref(false)
const categoryCreating = ref(false)
const editingCategoryName = ref('')
const categorySaving = ref(false)

// Выбранная категория для редактирования прав
const selectedCategoryId = ref<string | null>(null)
const categoryPerms = ref<CategoryPermissions | null>(null)
const categoryPermsLoading = ref(false)
const catPermsSaving = ref(false)
const catPermsSaved = ref(false)
const catDraftRoleIds = ref<Set<string>>(new Set())
const catDraftUserIds = ref<Set<string>>(new Set())

const categories = computed<ChannelCategory[]>(() => channelsStore.categories)
const selectedCategory = computed(() => categories.value.find(c => c.id === selectedCategoryId.value) ?? null)

async function selectCategory(cat: ChannelCategory) {
  selectedCategoryId.value = cat.id
  editingCategoryName.value = cat.name
  await fetchCategoryPerms(cat.id)
}

async function fetchCategoryPerms(categoryId: string) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  categoryPermsLoading.value = true
  try {
    categoryPerms.value = await categoriesApi.getPermissions(wsId, categoryId)
    catDraftRoleIds.value = new Set(categoryPerms.value.roles.map(r => r.role_id))
    catDraftUserIds.value = new Set(categoryPerms.value.users.map(u => u.user_id))
  } catch {
    categoryPerms.value = null
  } finally {
    categoryPermsLoading.value = false
  }
}

function toggleCatRole(roleId: string) {
  if (catDraftRoleIds.value.has(roleId)) catDraftRoleIds.value.delete(roleId)
  else catDraftRoleIds.value.add(roleId)
}

function toggleCatUser(userId: string) {
  if (catDraftUserIds.value.has(userId)) catDraftUserIds.value.delete(userId)
  else catDraftUserIds.value.add(userId)
}

async function saveCategoryPerms() {
  const wsId = workspaceStore.currentWorkspaceId
  const catId = selectedCategoryId.value
  if (!wsId || !catId || !categoryPerms.value) return
  catPermsSaving.value = true
  try {
    const currentRoles = new Set(categoryPerms.value.roles.map(r => r.role_id))
    const currentUsers = new Set(categoryPerms.value.users.map(u => u.user_id))

    const rolesToAdd = [...catDraftRoleIds.value].filter(id => !currentRoles.has(id))
    const rolesToRemove = [...currentRoles].filter(id => !catDraftRoleIds.value.has(id))
    const usersToAdd = [...catDraftUserIds.value].filter(id => !currentUsers.has(id))
    const usersToRemove = [...currentUsers].filter(id => !catDraftUserIds.value.has(id))

    await Promise.all([
      ...rolesToAdd.map(id => categoriesApi.addRole(wsId, catId, id)),
      ...rolesToRemove.map(id => categoriesApi.removeRole(wsId, catId, id)),
      ...usersToAdd.map(id => categoriesApi.addUser(wsId, catId, id)),
      ...usersToRemove.map(id => categoriesApi.removeUser(wsId, catId, id)),
    ])

    await fetchCategoryPerms(catId)
    catPermsSaved.value = true
    setTimeout(() => { catPermsSaved.value = false }, 2000)
  } finally {
    catPermsSaving.value = false
  }
}

async function saveCategoryName() {
  const wsId = workspaceStore.currentWorkspaceId
  const catId = selectedCategoryId.value
  if (!wsId || !catId || !editingCategoryName.value.trim()) return
  categorySaving.value = true
  try {
    await categoriesApi.update(wsId, catId, { name: editingCategoryName.value.trim() })
    await channelsStore.fetchCategories(wsId)
  } finally {
    categorySaving.value = false
  }
}

async function toggleCategoryPrivacy(cat: ChannelCategory) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  await categoriesApi.update(wsId, cat.id, { is_private: !cat.is_private })
  await channelsStore.fetchCategories(wsId)
  if (selectedCategoryId.value === cat.id && !cat.is_private) {
    await fetchCategoryPerms(cat.id)
  }
}

async function createCategory() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId || !newCategoryName.value.trim()) return
  categoryCreating.value = true
  try {
    await categoriesApi.create(wsId, { name: newCategoryName.value.trim(), is_private: newCategoryPrivate.value })
    newCategoryName.value = ''
    newCategoryPrivate.value = false
    await channelsStore.fetchCategories(wsId)
  } finally {
    categoryCreating.value = false
  }
}

async function deleteCategory(cat: ChannelCategory) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  if (!confirm(`Удалить категорию «${cat.name}»? Каналы останутся, но потеряют категорию.`)) return
  await categoriesApi.delete(wsId, cat.id)
  if (selectedCategoryId.value === cat.id) selectedCategoryId.value = null
  await channelsStore.fetchCategories(wsId)
}

// --- Members tab ---
const members = computed<WorkspaceMember[]>(() => {
  const wsId = workspaceStore.currentWorkspaceId
  return wsId ? (workspaceStore.membersMap[wsId] ?? []) : []
})

const currentUserRole = computed<WorkspaceMember['role'] | null>(() => {
  const me = members.value.find(m => m.user_id === authStore.user?.id)
  return me?.role ?? null
})

const canManageMembers = computed(() =>
  currentUserRole.value === 'owner' || currentUserRole.value === 'admin'
)

const memberActionLoading = ref<string | null>(null)
const memberActionError = ref('')

const systemRoleOptions: { value: WorkspaceMember['role']; label: string }[] = [
  { value: 'admin', label: 'Администратор' },
  { value: 'member', label: 'Участник' },
]

const roleLabels: Record<string, string> = {
  owner: 'Владелец',
  admin: 'Администратор',
  member: 'Участник',
}

function canEdit(member: WorkspaceMember) {
  if (!canManageMembers.value) return false
  if (member.role === 'owner') return false
  if (currentUserRole.value === 'admin' && member.role === 'admin') return false
  if (member.user_id === authStore.user?.id) return false
  return true
}

async function changeRole(member: WorkspaceMember, role: WorkspaceMember['role']) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  memberActionLoading.value = member.user_id
  memberActionError.value = ''
  try {
    await workspaceStore.updateMember(wsId, member.user_id, { role })
  } catch {
    memberActionError.value = 'Не удалось изменить роль'
  } finally {
    memberActionLoading.value = null
  }
}

async function kickMember(member: WorkspaceMember) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  if (!confirm(`Исключить ${member.display_name} из воркспейса?`)) return
  memberActionLoading.value = member.user_id
  memberActionError.value = ''
  try {
    await workspaceStore.kickMember(wsId, member.user_id)
  } catch {
    memberActionError.value = 'Не удалось исключить участника'
  } finally {
    memberActionLoading.value = null
  }
}

function getInitials(name: string) {
  return name.split(' ').map(p => p[0]).join('').toUpperCase().slice(0, 2)
}

// --- Roles tab ---
const roles = ref<WorkspaceRole[]>([])
const rolesLoading = ref(false)
const rolesError = ref('')
const selectedRole = ref<WorkspaceRole | null>(null)

// Редактирование выбранной роли
const editName = ref('')
const editColor = ref('')
const editPerms = ref<RolePermissions>(defaultPermissions())
const roleSaving = ref(false)
const roleSaveSuccess = ref(false)
const roleDeleteLoading = ref(false)

// Создание новой роли
const showCreateForm = ref(false)
const newRoleName = ref('')
const newRoleColor = ref('#5865f2')
const creating = ref(false)

// Назначение кастомной роли участнику прямо в списке участников
const memberRoles = ref<Record<string, WorkspaceRole[]>>({})

const permissionLabels: { key: keyof RolePermissions; label: string; description: string }[] = [
  { key: 'view_channels', label: 'Просматривать каналы', description: 'Видеть публичные каналы воркспейса' },
  { key: 'send_messages', label: 'Отправлять сообщения', description: 'Писать сообщения в каналах' },
  { key: 'attach_files', label: 'Прикреплять файлы', description: 'Загружать файлы и изображения' },
  { key: 'mention_everyone', label: 'Упоминать @everyone', description: 'Пинговать всех участников' },
  { key: 'manage_messages', label: 'Управлять сообщениями', description: 'Удалять и редактировать чужие сообщения' },
  { key: 'manage_channels', label: 'Управлять каналами', description: 'Создавать, редактировать и удалять каналы' },
  { key: 'manage_members', label: 'Управлять участниками', description: 'Исключать участников и назначать роли' },
  { key: 'manage_roles', label: 'Управлять ролями', description: 'Создавать и редактировать роли ниже своей' },
  { key: 'manage_workspace', label: 'Управлять воркспейсом', description: 'Редактировать название и описание' },
  { key: 'view_audit_log', label: 'Просматривать журнал', description: 'Видеть историю действий в воркспейсе' },
]

watch(activeTab, async (tab) => {
  if (tab === 'roles') await fetchRoles()
  if (tab === 'members') {
    // Подгружаем кастомные роли для отображения бейджей
    const wsId = workspaceStore.currentWorkspaceId
    if (!wsId) return
    if (!roles.value.length) await fetchRoles()
    for (const member of members.value) {
      if (!memberRoles.value[member.user_id]) {
        loadMemberRoles(member.user_id)
      }
    }
  }
})

async function fetchRoles() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  rolesLoading.value = true
  rolesError.value = ''
  try {
    roles.value = await rolesApi.list(wsId)
    // Если ничего не выбрано — выбираем @everyone
    if (!selectedRole.value && roles.value.length) {
      const everyone = roles.value.find(r => r.name === '@everyone')
      selectRole(everyone ?? roles.value[0])
    } else if (selectedRole.value) {
      // Обновляем выбранную роль из списка
      const updated = roles.value.find(r => r.id === selectedRole.value!.id)
      if (updated) selectRole(updated)
    }
  } catch {
    rolesError.value = 'Не удалось загрузить роли'
  } finally {
    rolesLoading.value = false
  }
}

function selectRole(role: WorkspaceRole) {
  selectedRole.value = role
  editName.value = role.name
  editColor.value = role.color
  editPerms.value = { ...role.permissions }
  roleSaveSuccess.value = false
  showCreateForm.value = false
}

async function saveRole() {
  if (!selectedRole.value || !workspaceStore.currentWorkspaceId) return
  roleSaving.value = true
  rolesError.value = ''
  try {
    const wsId = workspaceStore.currentWorkspaceId
    if (selectedRole.value.name === '@everyone') {
      await rolesApi.updateEveryone(wsId, editPerms.value)
    } else {
      await rolesApi.update(wsId, selectedRole.value.id, {
        name: selectedRole.value.is_system ? undefined : editName.value.trim() || undefined,
        color: selectedRole.value.is_system ? undefined : editColor.value || undefined,
        permissions: editPerms.value,
      })
    }
    roleSaveSuccess.value = true
    setTimeout(() => { roleSaveSuccess.value = false }, 2000)
    await fetchRoles()
  } catch {
    rolesError.value = 'Не удалось сохранить роль'
  } finally {
    roleSaving.value = false
  }
}

async function deleteRole() {
  if (!selectedRole.value || !workspaceStore.currentWorkspaceId) return
  if (!confirm(`Удалить роль «${selectedRole.value.name}»?`)) return
  roleDeleteLoading.value = true
  try {
    await rolesApi.delete(workspaceStore.currentWorkspaceId, selectedRole.value.id)
    selectedRole.value = null
    await fetchRoles()
    if (roles.value.length) selectRole(roles.value[0])
  } catch {
    rolesError.value = 'Не удалось удалить роль'
  } finally {
    roleDeleteLoading.value = false
  }
}

async function createRole() {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId || !newRoleName.value.trim()) return
  creating.value = true
  rolesError.value = ''
  try {
    const role = await rolesApi.create(wsId, {
      name: newRoleName.value.trim(),
      color: newRoleColor.value,
      permissions: defaultPermissions(),
    })
    newRoleName.value = ''
    newRoleColor.value = '#5865f2'
    showCreateForm.value = false
    await fetchRoles()
    selectRole(role)
  } catch {
    rolesError.value = 'Не удалось создать роль'
  } finally {
    creating.value = false
  }
}

// Назначение роли участнику
async function assignMemberRole(member: WorkspaceMember, roleId: string) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  try {
    await rolesApi.assignRole(wsId, member.user_id, roleId)
    await loadMemberRoles(member.user_id)
  } catch {
    memberActionError.value = 'Не удалось назначить роль'
  }
}

async function revokeMemberRole(member: WorkspaceMember, roleId: string) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  try {
    await rolesApi.revokeRole(wsId, member.user_id, roleId)
    await loadMemberRoles(member.user_id)
  } catch {
    memberActionError.value = 'Не удалось снять роль'
  }
}

async function loadMemberRoles(userId: string) {
  const wsId = workspaceStore.currentWorkspaceId
  if (!wsId) return
  const r = await rolesApi.getMemberRoles(wsId, userId)
  memberRoles.value[userId] = r
}

// Кастомные роли (не системные owner/admin/@everyone) для назначения участникам
const customRoles = computed(() =>
  roles.value.filter(r => !r.is_system)
)

// Попап назначения роли
const rolePopupUserId = ref<string | null>(null)
// Участник у которого открыт селект системной роли
const selectOpenMemberId = ref<string | null>(null)

// Любое из двух состояний — держим actions видимыми
function isActionsVisible(userId: string): boolean {
  return rolePopupUserId.value === userId || selectOpenMemberId.value === userId
}

function toggleRolePopup(userId: string) {
  rolePopupUserId.value = rolePopupUserId.value === userId ? null : userId
}

function closeRolePopup() {
  rolePopupUserId.value = null
}

function hasRole(member: WorkspaceMember, roleId: string): boolean {
  return (memberRoles.value[member.user_id] ?? []).some(r => r.id === roleId)
}

async function toggleMemberRole(member: WorkspaceMember, roleId: string) {
  if (hasRole(member, roleId)) {
    await revokeMemberRole(member, roleId)
  } else {
    await assignMemberRole(member, roleId)
  }
}
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
        v-if="props.open"
        class="fixed inset-0 z-50 flex bg-dark-950/80 backdrop-blur-sm"
        @click.self="emit('close')"
        @click="closeRolePopup"
      >
        <!-- Settings layout: sidebar + content -->
        <div class="m-auto flex w-full max-w-4xl h-[640px] card overflow-hidden p-0">
          <!-- Sidebar -->
          <div class="w-52 bg-dark-900 border-r border-dark-700 flex flex-col p-3 shrink-0">
            <p class="text-xs font-semibold text-dark-400 uppercase px-2 mb-2 truncate">
              {{ workspaceStore.currentWorkspace?.name }}
            </p>

            <button
              v-for="tab in [
                { id: 'general', label: 'Основные' },
                { id: 'roles', label: 'Роли' },
                { id: 'members', label: 'Участники' },
                { id: 'categories', label: 'Категории' },
              ]"
              :key="tab.id"
              class="w-full text-left px-2 py-1.5 rounded-md text-sm transition-colors"
              :class="activeTab === tab.id
                ? 'bg-dark-700 text-white'
                : 'text-dark-300 hover:bg-dark-800 hover:text-dark-100'"
              @click="activeTab = tab.id as Tab"
            >
              {{ tab.label }}
            </button>

            <div class="mt-auto pt-3 border-t border-dark-700">
              <button
                class="w-full text-left px-2 py-1.5 rounded-md text-sm text-dark-400 hover:bg-dark-800 transition-colors"
                @click="emit('close')"
              >
                Закрыть
              </button>
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-hidden flex flex-col">
            <!-- ===== General ===== -->
            <div
              v-if="activeTab === 'general'"
              class="flex-1 overflow-y-auto p-6 space-y-6"
            >
              <h2 class="text-lg font-semibold text-white">
                Основные настройки
              </h2>

              <div class="space-y-4">
                <!-- Иконка воркспейса -->
                <div>
                  <label class="block text-sm font-medium text-dark-300 mb-2">Иконка воркспейса</label>
                  <div class="flex items-center gap-4">
                    <div
                      class="relative w-20 h-20 rounded-2xl overflow-hidden bg-dark-800 border-2 border-dark-700 cursor-pointer group shrink-0"
                      :class="{ 'opacity-60': iconUploading }"
                      @click="iconInputRef?.click()"
                    >
                      <img
                        v-if="workspaceStore.currentWorkspace?.icon_url"
                        :src="workspaceStore.currentWorkspace.icon_url"
                        class="w-full h-full object-cover"
                        alt="icon"
                      >
                      <div
                        v-else
                        class="w-full h-full flex items-center justify-center text-2xl font-bold text-dark-400 select-none"
                      >
                        {{ workspaceStore.currentWorkspace?.name?.charAt(0)?.toUpperCase() }}
                      </div>
                      <div class="absolute inset-0 bg-black/50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                        <svg
                          v-if="!iconUploading"
                          class="w-6 h-6 text-white"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                          />
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"
                          />
                        </svg>
                        <svg
                          v-else
                          class="w-5 h-5 text-white animate-spin"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                          />
                          <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8v8H4z"
                          />
                        </svg>
                      </div>
                    </div>
                    <div class="text-xs text-dark-400 leading-relaxed">
                      <p>Нажмите на иконку чтобы загрузить изображение</p>
                      <p class="mt-1">
                        JPG, PNG, GIF · до 10 МБ
                      </p>
                    </div>
                    <input
                      ref="iconInputRef"
                      type="file"
                      accept="image/*"
                      class="hidden"
                      @change="onIconChange"
                    >
                  </div>
                </div>

                <div>
                  <label class="block text-sm font-medium text-dark-300 mb-1.5">Название</label>
                  <Input
                    v-model="name"
                    placeholder="Название воркспейса"
                    maxlength="100"
                  />
                </div>

                <div>
                  <label class="block text-sm font-medium text-dark-300 mb-1.5">Описание</label>
                  <textarea
                    v-model="description"
                    rows="3"
                    maxlength="500"
                    placeholder="Краткое описание воркспейса..."
                    class="w-full px-3 py-2 bg-dark-900 border border-dark-700 rounded-lg text-dark-100 placeholder-dark-500 focus:border-atlas-500 focus:outline-none resize-none text-sm"
                  />
                  <p class="text-xs text-dark-500 mt-1 text-right">
                    {{ description.length }}/500
                  </p>
                </div>
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
                @click="saveGeneral"
              >
                Сохранить
              </Button>

              <div
                v-if="currentUserRole === 'owner'"
                class="mt-8 pt-6 border-t border-dark-700"
              >
                <h3 class="text-sm font-semibold text-red-400 mb-3">
                  Опасная зона
                </h3>
                <div class="p-4 border border-red-500/20 rounded-lg flex items-center justify-between gap-4">
                  <div>
                    <p class="text-sm font-medium text-white">
                      Удалить воркспейс
                    </p>
                    <p class="text-xs text-dark-400 mt-0.5">
                      Все каналы и сообщения будут удалены навсегда
                    </p>
                  </div>
                  <button
                    class="px-3 py-1.5 text-sm font-medium text-red-400 border border-red-500/30 rounded-lg hover:bg-red-500/10 transition-colors shrink-0"
                    @click="deleteWorkspace"
                  >
                    Удалить
                  </button>
                </div>
              </div>
            </div>

            <!-- ===== Roles ===== -->
            <div
              v-else-if="activeTab === 'roles'"
              class="flex-1 flex overflow-hidden"
            >
              <!-- Role list -->
              <div class="w-48 border-r border-dark-700 flex flex-col bg-dark-900/50 shrink-0">
                <div class="p-3 border-b border-dark-700 flex items-center justify-between">
                  <span class="text-xs font-semibold text-dark-400 uppercase">Роли</span>
                  <button
                    v-if="canManageMembers"
                    class="p-1 text-dark-400 hover:text-white transition-colors"
                    title="Создать роль"
                    @click="showCreateForm = true; selectedRole = null"
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
                  </button>
                </div>

                <div
                  v-if="rolesLoading"
                  class="p-3 text-xs text-dark-500"
                >
                  Загрузка...
                </div>
                <div
                  v-else
                  class="flex-1 overflow-y-auto p-2 space-y-0.5"
                >
                  <button
                    v-for="role in roles"
                    :key="role.id"
                    class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left"
                    :class="selectedRole?.id === role.id
                      ? 'bg-dark-700 text-white'
                      : 'text-dark-300 hover:bg-dark-800'"
                    @click="selectRole(role)"
                  >
                    <!-- Color dot -->
                    <span
                      class="w-3 h-3 rounded-full shrink-0"
                      :style="{ backgroundColor: role.color }"
                    />
                    <span class="truncate">{{ role.name }}</span>
                    <span
                      v-if="role.is_system"
                      class="ml-auto text-dark-600 text-xs"
                    >🔒</span>
                  </button>
                </div>
              </div>

              <!-- Role editor -->
              <div class="flex-1 overflow-y-auto p-5">
                <!-- Форма создания -->
                <div
                  v-if="showCreateForm"
                  class="space-y-4"
                >
                  <h3 class="text-base font-semibold text-white">
                    Новая роль
                  </h3>
                  <div>
                    <label class="block text-sm font-medium text-dark-300 mb-1.5">Название роли</label>
                    <Input
                      v-model="newRoleName"
                      placeholder="Например: Модератор"
                      maxlength="100"
                    />
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-dark-300 mb-1.5">Цвет</label>
                    <div class="flex items-center gap-3">
                      <input
                        v-model="newRoleColor"
                        type="color"
                        class="w-10 h-10 rounded cursor-pointer border border-dark-700 bg-transparent"
                      >
                      <span class="text-sm text-dark-300">{{ newRoleColor }}</span>
                    </div>
                  </div>
                  <div
                    v-if="rolesError"
                    class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
                  >
                    <p class="text-sm text-red-400">
                      {{ rolesError }}
                    </p>
                  </div>
                  <div class="flex gap-2">
                    <Button
                      :loading="creating"
                      :disabled="!newRoleName.trim()"
                      @click="createRole"
                    >
                      Создать
                    </Button>
                    <button
                      class="px-3 py-1.5 text-sm text-dark-400 hover:text-white transition-colors"
                      @click="showCreateForm = false"
                    >
                      Отмена
                    </button>
                  </div>
                </div>

                <!-- Редактор роли -->
                <div
                  v-else-if="selectedRole"
                  class="space-y-5"
                >
                  <!-- Заголовок роли -->
                  <div class="flex items-center gap-3">
                    <span
                      class="w-4 h-4 rounded-full shrink-0"
                      :style="{ backgroundColor: selectedRole.color }"
                    />
                    <h3 class="text-base font-semibold text-white">
                      {{ selectedRole.name }}
                    </h3>
                    <span
                      v-if="selectedRole.is_system"
                      class="text-xs px-2 py-0.5 bg-dark-800 text-dark-400 rounded"
                    >Системная</span>
                  </div>

                  <!-- Имя и цвет — только для кастомных -->
                  <template v-if="!selectedRole.is_system && canManageMembers">
                    <div class="grid grid-cols-2 gap-4">
                      <div>
                        <label class="block text-sm font-medium text-dark-300 mb-1.5">Название</label>
                        <Input
                          v-model="editName"
                          placeholder="Название роли"
                          maxlength="100"
                        />
                      </div>
                      <div>
                        <label class="block text-sm font-medium text-dark-300 mb-1.5">Цвет</label>
                        <div class="flex items-center gap-2">
                          <input
                            v-model="editColor"
                            type="color"
                            class="w-10 h-10 rounded cursor-pointer border border-dark-700 bg-transparent"
                          >
                          <span class="text-sm text-dark-400">{{ editColor }}</span>
                        </div>
                      </div>
                    </div>
                  </template>

                  <!-- Permissions -->
                  <div>
                    <p class="text-xs font-semibold text-dark-400 uppercase mb-3">
                      Права доступа
                    </p>
                    <div class="space-y-2">
                      <div
                        v-for="perm in permissionLabels"
                        :key="perm.key"
                        class="flex items-start gap-3 p-3 rounded-lg"
                        :class="editPerms[perm.key] ? 'bg-atlas-600/10 border border-atlas-600/20' : 'bg-dark-900 border border-dark-800'"
                      >
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium text-white">
                            {{ perm.label }}
                          </p>
                          <p class="text-xs text-dark-400 mt-0.5">
                            {{ perm.description }}
                          </p>
                        </div>
                        <!-- owner/admin всегда locked -->
                        <template v-if="selectedRole.name === 'owner' || selectedRole.name === 'admin'">
                          <span class="text-xs text-atlas-400 mt-0.5">Всегда ✓</span>
                        </template>
                        <template v-else>
                          <button
                            class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0 mt-0.5"
                            :class="editPerms[perm.key] ? 'bg-atlas-600' : 'bg-dark-700'"
                            :disabled="!canManageMembers"
                            @click="editPerms[perm.key] = !editPerms[perm.key]"
                          >
                            <span
                              class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
                              :class="editPerms[perm.key] ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                            />
                          </button>
                        </template>
                      </div>
                    </div>
                  </div>

                  <!-- Ошибки / успех -->
                  <div
                    v-if="rolesError"
                    class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
                  >
                    <p class="text-sm text-red-400">
                      {{ rolesError }}
                    </p>
                  </div>
                  <div
                    v-if="roleSaveSuccess"
                    class="p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg"
                  >
                    <p class="text-sm text-emerald-400">
                      Роль сохранена
                    </p>
                  </div>

                  <!-- Actions -->
                  <div
                    v-if="canManageMembers"
                    class="flex items-center gap-3 pt-2"
                  >
                    <Button
                      :loading="roleSaving"
                      @click="saveRole"
                    >
                      Сохранить
                    </Button>
                    <button
                      v-if="!selectedRole.is_system"
                      :disabled="roleDeleteLoading"
                      class="px-3 py-1.5 text-sm text-red-400 border border-red-500/30 rounded-lg hover:bg-red-500/10 transition-colors"
                      @click="deleteRole"
                    >
                      Удалить роль
                    </button>
                  </div>
                </div>

                <div
                  v-else
                  class="flex items-center justify-center h-full text-dark-500 text-sm"
                >
                  Выберите роль из списка
                </div>
              </div>
            </div>

            <!-- ===== Members ===== -->
            <div
              v-else-if="activeTab === 'members'"
              class="flex-1 overflow-y-auto p-6 space-y-4"
            >
              <div class="flex items-center justify-between">
                <h2 class="text-lg font-semibold text-white">
                  Участники
                  <span class="ml-2 text-sm font-normal text-dark-400">{{ members.length }}</span>
                </h2>
              </div>

              <div
                v-if="memberActionError"
                class="p-3 bg-red-500/10 border border-red-500/20 rounded-lg"
              >
                <p class="text-sm text-red-400">
                  {{ memberActionError }}
                </p>
              </div>

              <div class="space-y-1">
                <div
                  v-for="member in members"
                  :key="member.user_id"
                  class="flex items-center gap-3 px-3 py-2.5 rounded-lg group"
                  :class="isActionsVisible(member.user_id) ? 'bg-dark-800' : 'hover:bg-dark-800'"
                >
                  <!-- Avatar -->
                  <div class="w-9 h-9 rounded-full bg-atlas-600 flex items-center justify-center text-white text-sm font-semibold shrink-0">
                    <img
                      v-if="member.avatar_url"
                      :src="member.avatar_url"
                      :alt="member.display_name"
                      class="w-full h-full rounded-full object-cover"
                    >
                    <span v-else>{{ getInitials(member.display_name || '?') }}</span>
                  </div>

                  <!-- Name -->
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-white truncate">
                      {{ member.nickname || member.display_name }}
                      <span
                        v-if="member.nickname"
                        class="text-dark-400 font-normal text-xs ml-1"
                      >({{ member.display_name }})</span>
                    </p>
                    <div class="flex items-center gap-1.5 flex-wrap mt-0.5">
                      <!-- Системная роль -->
                      <span class="text-xs text-dark-400">{{ roleLabels[member.role] ?? member.role }}</span>
                      <!-- Кастомные роли -->
                      <span
                        v-for="r in (memberRoles[member.user_id] ?? [])"
                        :key="r.id"
                        class="inline-flex items-center gap-1 text-xs px-1.5 py-0.5 rounded-full border"
                        :style="{ borderColor: r.color + '60', color: r.color }"
                      >
                        <span
                          class="w-1.5 h-1.5 rounded-full"
                          :style="{ backgroundColor: r.color }"
                        />
                        {{ r.name }}
                        <button
                          v-if="canManageMembers"
                          class="ml-0.5 opacity-60 hover:opacity-100"
                          @click.stop="revokeMemberRole(member, r.id)"
                        >×</button>
                      </span>
                    </div>
                  </div>

                  <!-- Actions -->
                  <div
                    v-if="canEdit(member)"
                    class="flex items-center gap-2 transition-opacity"
                    :class="isActionsVisible(member.user_id) ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'"
                  >
                    <!-- Назначить кастомную роль — кнопка с dropdown -->
                    <div
                      v-if="customRoles.length"
                      class="relative"
                    >
                      <button
                        class="flex items-center gap-1 px-2 py-1 rounded text-xs border transition-colors"
                        :class="rolePopupUserId === member.user_id
                          ? 'border-atlas-500 text-atlas-300 bg-atlas-600/10'
                          : 'border-dark-600 text-dark-300 hover:border-dark-400 hover:text-dark-100'"
                        @click.stop="toggleRolePopup(member.user_id)"
                      >
                        <svg
                          class="w-3 h-3"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2.5"
                            d="M12 4v16m8-8H4"
                          />
                        </svg>
                        Роль
                      </button>

                      <!-- Dropdown с ролями -->
                      <div
                        v-if="rolePopupUserId === member.user_id"
                        class="absolute right-0 bottom-full mb-1 w-52 bg-dark-800 border border-dark-600 rounded-lg shadow-2xl py-1 z-50"
                        @click.stop
                      >
                        <p class="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-dark-400">
                          Роли
                        </p>
                        <button
                          v-for="role in customRoles"
                          :key="role.id"
                          class="w-full flex items-center gap-2.5 px-3 py-1.5 hover:bg-dark-700 transition-colors"
                          @click="toggleMemberRole(member, role.id)"
                        >
                          <span
                            class="w-3 h-3 rounded-sm border-2 flex items-center justify-center shrink-0 transition-colors"
                            :style="hasRole(member, role.id)
                              ? { backgroundColor: '#5865f2', borderColor: '#5865f2' }
                              : { borderColor: '#4c4880' }"
                          >
                            <svg
                              v-if="hasRole(member, role.id)"
                              class="w-2 h-2 text-white"
                              fill="currentColor"
                              viewBox="0 0 12 12"
                            >
                              <path
                                d="M10 3L5 8.5 2 5.5"
                                stroke="white"
                                stroke-width="1.5"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                fill="none"
                              />
                            </svg>
                          </span>
                          <span
                            class="w-2.5 h-2.5 rounded-full shrink-0"
                            :style="{ backgroundColor: role.color }"
                          />
                          <span class="text-sm text-dark-200 truncate">{{ role.name }}</span>
                        </button>
                      </div>
                    </div>

                    <!-- Системная роль -->
                    <Select
                      :model-value="member.role"
                      :options="systemRoleOptions"
                      :disabled="memberActionLoading === member.user_id"
                      :small="true"
                      class="w-32"
                      @open="selectOpenMemberId = member.user_id"
                      @close="selectOpenMemberId = null"
                      @update:model-value="changeRole(member, $event as WorkspaceMember['role'])"
                    />

                    <button
                      :disabled="memberActionLoading === member.user_id"
                      class="p-1 text-dark-400 hover:text-red-400 transition-colors"
                      title="Исключить"
                      @click="kickMember(member)"
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
                          d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
                        />
                      </svg>
                    </button>
                  </div>

                  <span
                    v-else-if="member.user_id === authStore.user?.id"
                    class="text-xs text-dark-500"
                  >Вы</span>
                </div>
              </div>
            </div>

            <!-- ===== Categories ===== -->
            <div
              v-else-if="activeTab === 'categories'"
              class="flex-1 overflow-hidden flex"
            >
              <!-- Left: список категорий -->
              <div class="w-52 border-r border-dark-700 flex flex-col bg-dark-900/50 shrink-0">
                <div class="p-3 border-b border-dark-700">
                  <span class="text-xs font-semibold text-dark-400 uppercase">Категории</span>
                </div>
                <div class="flex-1 overflow-y-auto p-2 space-y-0.5">
                  <button
                    v-for="cat in categories"
                    :key="cat.id"
                    class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-sm transition-colors text-left"
                    :class="selectedCategoryId === cat.id ? 'bg-dark-700 text-white' : 'text-dark-300 hover:bg-dark-800'"
                    @click="selectCategory(cat)"
                  >
                    <svg
                      v-if="cat.is_private"
                      class="w-3.5 h-3.5 shrink-0 text-dark-400"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fill-rule="evenodd"
                        d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
                        clip-rule="evenodd"
                      />
                    </svg>
                    <span class="truncate flex-1">{{ cat.name }}</span>
                  </button>
                  <p
                    v-if="categories.length === 0"
                    class="text-xs text-dark-500 text-center py-3"
                  >
                    Нет категорий
                  </p>
                </div>
                <!-- Создать категорию -->
                <div class="p-3 border-t border-dark-700 space-y-2">
                  <input
                    v-model="newCategoryName"
                    class="w-full px-2 py-1.5 bg-dark-800 border border-dark-700 rounded text-sm text-dark-100 placeholder-dark-500 focus:border-atlas-500 focus:outline-none"
                    placeholder="Новая категория..."
                    @keyup.enter="createCategory"
                  >
                  <div class="flex items-center justify-between">
                    <label class="flex items-center gap-1.5 cursor-pointer text-xs text-dark-400">
                      <input
                        v-model="newCategoryPrivate"
                        type="checkbox"
                        class="w-3.5 h-3.5 rounded"
                      >
                      Приватная
                    </label>
                    <Button
                      size="sm"
                      :loading="categoryCreating"
                      @click="createCategory"
                    >
                      Создать
                    </Button>
                  </div>
                </div>
              </div>

              <!-- Right: редактор выбранной категории -->
              <div class="flex-1 overflow-y-auto">
                <div
                  v-if="!selectedCategory"
                  class="flex items-center justify-center h-full text-dark-500 text-sm"
                >
                  Выберите категорию для настройки
                </div>

                <div
                  v-else
                  class="p-6 space-y-6"
                >
                  <div class="flex items-start justify-between gap-4">
                    <div class="flex-1 space-y-3">
                      <h3 class="text-base font-semibold text-white">
                        {{ selectedCategory.name }}
                      </h3>
                      <!-- Переименование -->
                      <div class="flex gap-2">
                        <input
                          v-model="editingCategoryName"
                          class="flex-1 px-3 py-1.5 bg-dark-900 border border-dark-700 rounded-lg text-sm text-dark-100 placeholder-dark-500 focus:border-atlas-500 focus:outline-none"
                          placeholder="Название категории"
                          @keyup.enter="saveCategoryName"
                        >
                        <button
                          class="px-3 py-1.5 text-sm bg-atlas-600 hover:bg-atlas-500 text-white rounded-lg transition-colors"
                          :disabled="categorySaving"
                          @click="saveCategoryName"
                        >
                          {{ categorySaving ? '...' : 'OK' }}
                        </button>
                      </div>
                    </div>
                    <button
                      class="p-1.5 text-dark-400 hover:text-red-400 transition-colors mt-0.5 shrink-0"
                      title="Удалить категорию"
                      @click="deleteCategory(selectedCategory)"
                    >
                      <svg
                        class="w-5 h-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                        />
                      </svg>
                    </button>
                  </div>

                  <!-- Приватность категории -->
                  <div
                    class="flex items-center justify-between p-3 rounded-lg cursor-pointer select-none"
                    :class="selectedCategory.is_private ? 'bg-atlas-600/10 border border-atlas-600/30' : 'bg-dark-800 border border-dark-700'"
                    @click="toggleCategoryPrivacy(selectedCategory)"
                  >
                    <div class="flex items-center gap-2">
                      <svg
                        class="w-4 h-4 text-dark-400 shrink-0"
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
                        <p class="text-sm font-medium text-dark-100">
                          Приватная категория
                        </p>
                        <p class="text-xs text-dark-500">
                          Только выбранные роли и участники
                        </p>
                      </div>
                    </div>
                    <div
                      class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0"
                      :class="selectedCategory.is_private ? 'bg-atlas-600' : 'bg-dark-600'"
                    >
                      <span
                        class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
                        :class="selectedCategory.is_private ? 'translate-x-[18px]' : 'translate-x-[3px]'"
                      />
                    </div>
                  </div>

                  <!-- Права доступа — только если категория приватная -->
                  <div
                    v-if="selectedCategory.is_private"
                    class="space-y-4"
                  >
                    <div class="flex items-center justify-between">
                      <h4 class="text-sm font-semibold text-dark-200">
                        Права доступа
                      </h4>
                      <p class="text-xs text-dark-500">
                        Права применяются ко всем каналам внутри
                      </p>
                    </div>

                    <div
                      v-if="categoryPermsLoading"
                      class="text-xs text-dark-500 py-2"
                    >
                      Загрузка...
                    </div>

                    <div
                      v-else
                      class="space-y-4"
                    >
                      <!-- Роли -->
                      <div>
                        <p class="text-xs font-semibold text-dark-400 uppercase mb-2">
                          Роли
                        </p>
                        <div class="space-y-1">
                          <label
                            v-for="role in roles.filter(r => !r.is_system)"
                            :key="role.id"
                            class="flex items-center gap-2.5 px-2 py-1.5 rounded-lg hover:bg-dark-800 cursor-pointer"
                          >
                            <span
                              class="w-3 h-3 rounded-sm border-2 flex items-center justify-center shrink-0 transition-colors"
                              :style="catDraftRoleIds.has(role.id)
                                ? { backgroundColor: '#5865f2', borderColor: '#5865f2' }
                                : { borderColor: '#4c4880' }"
                              @click.prevent="toggleCatRole(role.id)"
                            >
                              <svg
                                v-if="catDraftRoleIds.has(role.id)"
                                class="w-2 h-2 text-white"
                                fill="currentColor"
                                viewBox="0 0 12 12"
                              >
                                <path
                                  d="M10 3L5 8.5 2 5.5"
                                  stroke="white"
                                  stroke-width="1.5"
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  fill="none"
                                />
                              </svg>
                            </span>
                            <span
                              class="w-2.5 h-2.5 rounded-full shrink-0"
                              :style="{ backgroundColor: role.color }"
                            />
                            <span class="text-sm text-dark-200">{{ role.name }}</span>
                          </label>
                        </div>
                      </div>

                      <!-- Участники -->
                      <div>
                        <p class="text-xs font-semibold text-dark-400 uppercase mb-2">
                          Участники
                        </p>
                        <div class="space-y-1">
                          <label
                            v-for="member in members.filter(m => m.role !== 'owner' && m.role !== 'admin')"
                            :key="member.user_id"
                            class="flex items-center gap-2.5 px-2 py-1.5 rounded-lg hover:bg-dark-800 cursor-pointer"
                          >
                            <span
                              class="w-3 h-3 rounded-sm border-2 flex items-center justify-center shrink-0 transition-colors"
                              :style="catDraftUserIds.has(member.user_id)
                                ? { backgroundColor: '#5865f2', borderColor: '#5865f2' }
                                : { borderColor: '#4c4880' }"
                              @click.prevent="toggleCatUser(member.user_id)"
                            >
                              <svg
                                v-if="catDraftUserIds.has(member.user_id)"
                                class="w-2 h-2 text-white"
                                fill="currentColor"
                                viewBox="0 0 12 12"
                              >
                                <path
                                  d="M10 3L5 8.5 2 5.5"
                                  stroke="white"
                                  stroke-width="1.5"
                                  stroke-linecap="round"
                                  stroke-linejoin="round"
                                  fill="none"
                                />
                              </svg>
                            </span>
                            <div class="w-6 h-6 rounded-full bg-atlas-600 flex items-center justify-center text-[10px] text-white font-semibold shrink-0">
                              <img
                                v-if="member.avatar_url"
                                :src="member.avatar_url"
                                class="w-full h-full rounded-full object-cover"
                                alt=""
                              >
                              <span v-else>{{ getInitials(member.display_name || '?') }}</span>
                            </div>
                            <span class="text-sm text-dark-200 truncate">{{ member.nickname || member.display_name }}</span>
                          </label>
                        </div>
                      </div>

                      <div class="flex items-center gap-3 pt-2">
                        <Button
                          :loading="catPermsSaving"
                          @click="saveCategoryPerms"
                        >
                          Сохранить
                        </Button>
                        <span
                          v-if="catPermsSaved"
                          class="text-xs text-emerald-400"
                        >Сохранено ✓</span>
                      </div>
                    </div>
                  </div>

                  <!-- Каналы в категории -->
                  <div class="pt-4 border-t border-dark-700">
                    <p class="text-xs font-semibold text-dark-400 uppercase mb-2">
                      Каналов в категории: {{ channelsStore.channels.filter(ch => ch.category_id === selectedCategoryId).length }}
                    </p>
                    <div class="space-y-0.5">
                      <div
                        v-for="ch in channelsStore.channels.filter(c => c.category_id === selectedCategoryId)"
                        :key="ch.id"
                        class="flex items-center gap-2 px-2 py-1 text-sm text-dark-400"
                      >
                        <svg
                          class="w-3.5 h-3.5 shrink-0"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14"
                          />
                        </svg>
                        <span class="truncate">{{ ch.name }}</span>
                      </div>
                      <p
                        v-if="channelsStore.channels.filter(c => c.category_id === selectedCategoryId).length === 0"
                        class="text-xs text-dark-600 px-2"
                      >
                        Нет каналов
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
