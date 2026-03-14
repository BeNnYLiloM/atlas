import apiClient from './client'
import type { Project, ProjectCreate, ProjectUpdate, ProjectMember, ProjectMemberAdd, ApiResponse } from '@/types'

export const projectsApi = {
  async list(workspaceId: string): Promise<Project[]> {
    const response = await apiClient.get<ApiResponse<Project[]>>(`/workspaces/${workspaceId}/projects`)
    return response.data.data ?? []
  },

  async get(id: string): Promise<Project> {
    const response = await apiClient.get<ApiResponse<Project>>(`/projects/${id}`)
    return response.data.data
  },

  async create(workspaceId: string, data: ProjectCreate): Promise<Project> {
    const response = await apiClient.post<ApiResponse<Project>>(`/workspaces/${workspaceId}/projects`, data)
    return response.data.data
  },

  async update(id: string, data: ProjectUpdate): Promise<Project> {
    const response = await apiClient.patch<ApiResponse<Project>>(`/projects/${id}`, data)
    return response.data.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/projects/${id}?force=true`)
  },

  async archive(id: string): Promise<void> {
    await apiClient.post(`/projects/${id}/archive`)
  },

  async unarchive(id: string): Promise<void> {
    await apiClient.delete(`/projects/${id}/archive`)
  },

  async getMembers(id: string): Promise<ProjectMember[]> {
    const response = await apiClient.get<ApiResponse<ProjectMember[]>>(`/projects/${id}/members`)
    return response.data.data ?? []
  },

  async addMember(id: string, data: ProjectMemberAdd): Promise<void> {
    await apiClient.post(`/projects/${id}/members`, data)
  },

  async removeMember(id: string, userId: string): Promise<void> {
    await apiClient.delete(`/projects/${id}/members/${userId}`)
  },

  async setLead(id: string, userId: string): Promise<void> {
    await apiClient.post(`/projects/${id}/members/${userId}/lead`)
  },

  async unsetLead(id: string, userId: string): Promise<void> {
    await apiClient.delete(`/projects/${id}/members/${userId}/lead`)
  },

  async uploadIcon(id: string, file: File): Promise<Project> {
    const formData = new FormData()
    formData.append('icon', file)
    const response = await apiClient.post<ApiResponse<Project>>(`/projects/${id}/icon`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data.data
  },
}
