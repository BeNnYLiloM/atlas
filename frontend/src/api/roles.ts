import apiClient from './client'
import type { WorkspaceRole, WorkspaceRoleCreate, WorkspaceRoleUpdate, RolePermissions, ApiResponse } from '@/types'

export const rolesApi = {
  async list(workspaceId: string): Promise<WorkspaceRole[]> {
    const res = await apiClient.get<ApiResponse<WorkspaceRole[]>>(`/workspaces/${workspaceId}/roles`)
    return res.data.data ?? []
  },

  async create(workspaceId: string, input: WorkspaceRoleCreate): Promise<WorkspaceRole> {
    const res = await apiClient.post<ApiResponse<WorkspaceRole>>(`/workspaces/${workspaceId}/roles`, input)
    return res.data.data
  },

  async update(workspaceId: string, roleId: string, input: WorkspaceRoleUpdate): Promise<WorkspaceRole> {
    const res = await apiClient.patch<ApiResponse<WorkspaceRole>>(`/workspaces/${workspaceId}/roles/${roleId}`, input)
    return res.data.data
  },

  async updateEveryone(workspaceId: string, permissions: RolePermissions): Promise<WorkspaceRole> {
    const res = await apiClient.patch<ApiResponse<WorkspaceRole>>(`/workspaces/${workspaceId}/roles/everyone`, permissions)
    return res.data.data
  },

  async delete(workspaceId: string, roleId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/roles/${roleId}`)
  },

  async assignRole(workspaceId: string, userId: string, roleId: string): Promise<void> {
    await apiClient.post(`/workspaces/${workspaceId}/members/${userId}/roles`, { role_id: roleId })
  },

  async revokeRole(workspaceId: string, userId: string, roleId: string): Promise<void> {
    await apiClient.delete(`/workspaces/${workspaceId}/members/${userId}/roles/${roleId}`)
  },

  async getMemberRoles(workspaceId: string, userId: string): Promise<WorkspaceRole[]> {
    const res = await apiClient.get<ApiResponse<WorkspaceRole[]>>(`/workspaces/${workspaceId}/members/${userId}/roles`)
    return res.data.data ?? []
  },
}
