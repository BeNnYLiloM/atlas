import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { projectsApi } from '@/api'
import type { Project, ProjectCreate, ProjectUpdate, ProjectMember } from '@/types'

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<Project[]>([])
  const currentProjectId = ref<string | null>(null)
  const membersMap = ref<Record<string, ProjectMember[]>>({})
  const loading = ref(false)

  const currentProject = computed(() =>
    projects.value.find(p => p.id === currentProjectId.value) ?? null
  )

  const currentMembers = computed(() =>
    currentProjectId.value ? (membersMap.value[currentProjectId.value] ?? []) : []
  )

  async function fetchProjects(workspaceId: string) {
    loading.value = true
    try {
      projects.value = await projectsApi.list(workspaceId)
    } catch {
      projects.value = []
    } finally {
      loading.value = false
    }
  }

  async function createProject(workspaceId: string, data: ProjectCreate): Promise<Project> {
    const project = await projectsApi.create(workspaceId, data)
    projects.value.push(project)
    return project
  }

  async function updateProject(id: string, data: ProjectUpdate): Promise<Project> {
    const updated = await projectsApi.update(id, data)
    const idx = projects.value.findIndex(p => p.id === id)
    if (idx !== -1) projects.value[idx] = updated
    return updated
  }

  async function deleteProject(id: string): Promise<void> {
    await projectsApi.delete(id)
    projects.value = projects.value.filter(p => p.id !== id)
    if (currentProjectId.value === id) currentProjectId.value = null
  }

  async function archiveProject(id: string): Promise<void> {
    await projectsApi.archive(id)
    const idx = projects.value.findIndex(p => p.id === id)
    if (idx !== -1) projects.value[idx] = { ...projects.value[idx], is_archived: true }
  }

  async function unarchiveProject(id: string): Promise<void> {
    await projectsApi.unarchive(id)
    const idx = projects.value.findIndex(p => p.id === id)
    if (idx !== -1) projects.value[idx] = { ...projects.value[idx], is_archived: false }
  }

  async function fetchMembers(projectId: string): Promise<void> {
    membersMap.value[projectId] = await projectsApi.getMembers(projectId)
  }

  async function addMember(projectId: string, userId: string): Promise<void> {
    await projectsApi.addMember(projectId, { user_id: userId })
    await fetchMembers(projectId)
  }

  async function removeMember(projectId: string, userId: string): Promise<void> {
    await projectsApi.removeMember(projectId, userId)
    if (membersMap.value[projectId]) {
      membersMap.value[projectId] = membersMap.value[projectId].filter(m => m.user_id !== userId)
    }
  }

  async function setLead(projectId: string, userId: string, isLead: boolean): Promise<void> {
    if (isLead) {
      await projectsApi.setLead(projectId, userId)
    } else {
      await projectsApi.unsetLead(projectId, userId)
    }
    if (membersMap.value[projectId]) {
      const m = membersMap.value[projectId].find(m => m.user_id === userId)
      if (m) m.is_lead = isLead
    }
  }

  // WS event handlers

  function onProjectCreated(project: Project) {
    if (!projects.value.find(p => p.id === project.id)) {
      projects.value.push(project)
    }
  }

  function onProjectUpdated(project: Project) {
    const idx = projects.value.findIndex(p => p.id === project.id)
    if (idx !== -1) projects.value[idx] = project
  }

  function onProjectDeleted(payload: { project_id: string }) {
    projects.value = projects.value.filter(p => p.id !== payload.project_id)
    if (currentProjectId.value === payload.project_id) currentProjectId.value = null
  }

  function onMemberAdded(payload: { project_id: string; user_id: string }) {
    // При необходимости перезагружаем список участников
    if (membersMap.value[payload.project_id]) {
      fetchMembers(payload.project_id)
    }
  }

  function onMemberRemoved(payload: { project_id: string; user_id: string }) {
    if (membersMap.value[payload.project_id]) {
      membersMap.value[payload.project_id] = membersMap.value[payload.project_id]
        .filter(m => m.user_id !== payload.user_id)
    }
  }

  function reset() {
    projects.value = []
    currentProjectId.value = null
    membersMap.value = {}
  }

  return {
    projects,
    currentProjectId,
    currentProject,
    currentMembers,
    membersMap,
    loading,
    fetchProjects,
    createProject,
    updateProject,
    deleteProject,
    archiveProject,
    unarchiveProject,
    fetchMembers,
    addMember,
    removeMember,
    setLead,
    onProjectCreated,
    onProjectUpdated,
    onProjectDeleted,
    onMemberAdded,
    onMemberRemoved,
    reset,
  }
})
