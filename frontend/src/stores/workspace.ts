import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { workspacesApi } from '@/api'
import type { Workspace, WorkspaceCreate, WorkspaceUpdate, WorkspaceMember } from '@/types'

export const useWorkspaceStore = defineStore('workspace', () => {
  const workspaces = ref<Workspace[]>([])
  const currentWorkspaceId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Presence: userId -> status ('online' | 'away' | 'offline')
  const presenceMap = ref<Record<string, string>>({})

  // Members per workspace: workspaceId -> WorkspaceMember[]
  const membersMap = ref<Record<string, WorkspaceMember[]>>({})

  const currentWorkspace = computed(() =>
    workspaces.value.find(w => w.id === currentWorkspaceId.value) ?? null
  )

  async function fetchWorkspaces() {
    loading.value = true
    error.value = null
    try {
      const result = await workspacesApi.list()
      workspaces.value = result ?? []
      if (!currentWorkspaceId.value && workspaces.value.length > 0) {
        currentWorkspaceId.value = workspaces.value[0].id
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки воркспейсов'
    } finally {
      loading.value = false
    }
  }

  async function createWorkspace(data: WorkspaceCreate) {
    loading.value = true
    error.value = null
    try {
      const workspace = await workspacesApi.create(data)
      if (!workspaces.value) {
        workspaces.value = []
      }
      workspaces.value.push(workspace)
      currentWorkspaceId.value = workspace.id
      return workspace
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка создания воркспейса'
      throw e
    } finally {
      loading.value = false
    }
  }

  // Обновить настройки воркспейса через API
  async function updateWorkspace(id: string, data: WorkspaceUpdate) {
    const updated = await workspacesApi.update(id, data)
    const idx = workspaces.value.findIndex(w => w.id === id)
    if (idx !== -1) workspaces.value[idx] = updated
    return updated
  }

  // Применить WS-событие workspace_updated локально
  function applyWorkspaceUpdate(workspace: Workspace) {
    const idx = workspaces.value.findIndex(w => w.id === workspace.id)
    if (idx !== -1) workspaces.value[idx] = workspace
  }

  function setCurrentWorkspace(id: string) {
    currentWorkspaceId.value = id
  }

  // Presence методы
  function setPresence(userId: string, status: string) {
    presenceMap.value[userId] = status
  }

  function getPresence(userId: string): string {
    return presenceMap.value[userId] ?? 'offline'
  }

  async function fetchMembers(workspaceId: string) {
    try {
      const members = await workspacesApi.getMembers(workspaceId)
      membersMap.value[workspaceId] = members
    } catch (e) {
      console.error('Failed to fetch members:', e)
    }
  }

  // Обновить роль/никнейм через API + локально
  async function updateMember(workspaceId: string, userId: string, data: { role?: WorkspaceMember['role']; nickname?: string | null }) {
    await workspacesApi.updateMember(workspaceId, userId, data)
    applyMemberUpdate({ workspace_id: workspaceId, user_id: userId, ...data })
  }

  // Исключить участника через API + локально
  async function kickMember(workspaceId: string, userId: string) {
    await workspacesApi.removeMember(workspaceId, userId)
    applyMemberRemove({ workspace_id: workspaceId, user_id: userId })
  }

  // --- WS-event handlers (используются в websocket store) ---

  function addMember(data: { workspace_id: string; user_id: string; role: string }) {
    const members = membersMap.value[data.workspace_id]
    if (members) {
      const exists = members.find(m => m.user_id === data.user_id)
      if (!exists) {
        members.push({
          workspace_id: data.workspace_id,
          user_id: data.user_id,
          role: data.role as WorkspaceMember['role'],
          display_name: '',
          avatar_url: null,
          nickname: null,
          custom_roles: [],
        })
      }
    }
  }

  function applyMemberRemove(data: { workspace_id: string; user_id: string }) {
    const members = membersMap.value[data.workspace_id]
    if (members) {
      const index = members.findIndex(m => m.user_id === data.user_id)
      if (index !== -1) {
        members.splice(index, 1)
      }
    }
  }

  function applyMemberUpdate(data: { workspace_id: string; user_id: string; role?: string; nickname?: string | null }) {
    const members = membersMap.value[data.workspace_id]
    if (members) {
      const member = members.find(m => m.user_id === data.user_id)
      if (member) {
        if (data.role) member.role = data.role as WorkspaceMember['role']
        if (data.nickname !== undefined) member.nickname = data.nickname ?? null
      }
    }
  }

  function $reset() {
    workspaces.value = []
    currentWorkspaceId.value = null
    loading.value = false
    error.value = null
    presenceMap.value = {}
    membersMap.value = {}
  }

  return {
    workspaces,
    currentWorkspaceId,
    currentWorkspace,
    loading,
    error,
    presenceMap,
    membersMap,
    fetchWorkspaces,
    createWorkspace,
    updateWorkspace,
    applyWorkspaceUpdate,
    setCurrentWorkspace,
    setPresence,
    getPresence,
    fetchMembers,
    updateMember,
    kickMember,
    addMember,
    applyMemberRemove,
    applyMemberUpdate,
    $reset,
  }
})
