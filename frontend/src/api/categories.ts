import apiClient from './client'
import type { ChannelCategory, ChannelCategoryCreate, ChannelCategoryUpdate, CategoryPermissions, ApiResponse } from '@/types'

export const categoriesApi = {
  async list(workspaceId: string): Promise<ChannelCategory[]> {
    const response = await apiClient.get<ApiResponse<ChannelCategory[]>>(`/workspaces/${workspaceId}/categories`)
    return response.data.data ?? []
  },

  async create(workspaceId: string, data: ChannelCategoryCreate): Promise<ChannelCategory> {
    const response = await apiClient.post<ApiResponse<ChannelCategory>>(`/workspaces/${workspaceId}/categories`, data)
    return response.data.data
  },

  async update(workspaceId: string, categoryId: string, data: ChannelCategoryUpdate): Promise<ChannelCategory> {
    const response = await apiClient.patch<ApiResponse<ChannelCategory>>(`/workspaces/${workspaceId}/categories/${categoryId}`, data)
    return response.data.data
  },

  async delete(workspaceId: string, categoryId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/categories/${categoryId}`)
  },

  // --- Permissions ---
  async getPermissions(workspaceId: string, categoryId: string): Promise<CategoryPermissions> {
    const response = await apiClient.get<ApiResponse<CategoryPermissions>>(`/workspaces/${workspaceId}/categories/${categoryId}/permissions`)
    return response.data.data
  },

  async addRole(workspaceId: string, categoryId: string, roleId: string): Promise<void> {
    await apiClient.post(`/workspaces/${workspaceId}/categories/${categoryId}/permissions/roles`, { role_id: roleId })
  },

  async removeRole(workspaceId: string, categoryId: string, roleId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/categories/${categoryId}/permissions/roles/${roleId}`)
  },

  async addUser(workspaceId: string, categoryId: string, userId: string): Promise<void> {
    await apiClient.post(`/workspaces/${workspaceId}/categories/${categoryId}/permissions/users`, { user_id: userId })
  },

  async removeUser(workspaceId: string, categoryId: string, userId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/categories/${categoryId}/permissions/users/${userId}`)
  },
}
