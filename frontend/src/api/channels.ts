import apiClient from './client'
import type { Channel, ChannelCreate, ChannelUpdate, ChannelWithUnread, ChannelMemberInfo, ChannelPermissions, NotificationLevel, ApiResponse } from '@/types'

export const channelsApi = {
  async list(workspaceId: string): Promise<ChannelWithUnread[]> {
    const response = await apiClient.get<ApiResponse<ChannelWithUnread[]>>(`/workspaces/${workspaceId}/channels`)
    return response.data.data ?? []
  },

  async listByProject(projectId: string): Promise<ChannelWithUnread[]> {
    const response = await apiClient.get<ApiResponse<ChannelWithUnread[]>>(`/projects/${projectId}/channels`)
    return response.data.data ?? []
  },

  async get(channelId: string): Promise<Channel> {
    const response = await apiClient.get<ApiResponse<Channel>>(`/channels/${channelId}`)
    return response.data.data
  },

  async create(data: ChannelCreate): Promise<Channel> {
    const response = await apiClient.post<ApiResponse<Channel>>('/channels', data)
    return response.data.data
  },

  async update(channelId: string, data: ChannelUpdate): Promise<Channel> {
    const response = await apiClient.patch<ApiResponse<Channel>>(`/channels/${channelId}`, data)
    return response.data.data
  },

  async delete(channelId: string): Promise<void> {
    await apiClient.delete(`/channels/${channelId}`)
  },

  async markAsRead(channelId: string, messageId?: string): Promise<void> {
    await apiClient.post(`/channels/${channelId}/read`, messageId ? { message_id: messageId } : {})
  },

  async updateNotifications(channelId: string, level: NotificationLevel): Promise<void> {
    await apiClient.patch(`/channels/${channelId}/notifications`, { level })
  },

  async getMembers(channelId: string): Promise<ChannelMemberInfo[]> {
    const response = await apiClient.get<ApiResponse<ChannelMemberInfo[]>>(`/channels/${channelId}/members`)
    return response.data.data ?? []
  },

  async addMember(channelId: string, userId: string): Promise<void> {
    await apiClient.post(`/channels/${channelId}/members`, { user_id: userId })
  },

  async removeMember(channelId: string, userId: string): Promise<void> {
    await apiClient.delete(`/channels/${channelId}/members/${userId}`)
  },

  // Права доступа к каналу
  async getPermissions(channelId: string): Promise<ChannelPermissions> {
    const res = await apiClient.get<ApiResponse<ChannelPermissions>>(`/channels/${channelId}/permissions`)
    return res.data.data ?? { roles: [], users: [] }
  },

  async addRolePermission(channelId: string, roleId: string): Promise<void> {
    await apiClient.post(`/channels/${channelId}/permissions/roles`, { role_id: roleId })
  },

  async removeRolePermission(channelId: string, roleId: string): Promise<void> {
    await apiClient.delete(`/channels/${channelId}/permissions/roles/${roleId}`)
  },

  async addUserPermission(channelId: string, userId: string): Promise<void> {
    await apiClient.post(`/channels/${channelId}/permissions/users`, { user_id: userId })
  },

  async removeUserPermission(channelId: string, userId: string): Promise<void> {
    await apiClient.delete(`/channels/${channelId}/permissions/users/${userId}`)
  },

  async checkCanWrite(channelId: string): Promise<boolean> {
    const res = await apiClient.get<ApiResponse<{ can_write: boolean }>>(`/channels/${channelId}/can-write`)
    return res.data.data?.can_write ?? true
  },
}

