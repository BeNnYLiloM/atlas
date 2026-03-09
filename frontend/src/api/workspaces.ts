import apiClient from './client'
import type { Workspace, WorkspaceCreate, WorkspaceUpdate, WorkspaceMember, WorkspaceMemberUpdate, ApiResponse } from '@/types'

export const workspacesApi = {
  async list(): Promise<Workspace[]> {
    const response = await apiClient.get<ApiResponse<Workspace[]>>('/workspaces')
    return response.data.data ?? []
  },

  async get(id: string): Promise<Workspace> {
    const response = await apiClient.get<ApiResponse<Workspace>>(`/workspaces/${id}`)
    return response.data.data
  },

  async create(data: WorkspaceCreate): Promise<Workspace> {
    const response = await apiClient.post<ApiResponse<Workspace>>('/workspaces', data)
    return response.data.data
  },

  async update(id: string, data: WorkspaceUpdate): Promise<Workspace> {
    const response = await apiClient.patch<ApiResponse<Workspace>>(`/workspaces/${id}`, data)
    return response.data.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/workspaces/${id}`)
  },

  async getMembers(workspaceId: string): Promise<WorkspaceMember[]> {
    const response = await apiClient.get<ApiResponse<WorkspaceMember[]>>(`/workspaces/${workspaceId}/members`)
    return response.data.data
  },

  async addMember(workspaceId: string, userId: string, role: string): Promise<void> {
    await apiClient.post(`/workspaces/${workspaceId}/members`, { user_id: userId, role })
  },

  async updateMember(workspaceId: string, userId: string, data: WorkspaceMemberUpdate): Promise<void> {
    await apiClient.patch(`/workspaces/${workspaceId}/members/${userId}`, data)
  },

  async removeMember(workspaceId: string, userId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/members/${userId}`)
  },

  async uploadIcon(workspaceId: string, file: File): Promise<Workspace> {
    const formData = new FormData()
    formData.append('icon', file)
    const response = await apiClient.post<ApiResponse<Workspace>>(`/workspaces/${workspaceId}/icon`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data.data
  },
}

