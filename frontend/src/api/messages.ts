import apiClient from './client'
import type { Message, MessageCreate, MessageUpdate, ApiResponse } from '@/types'

export const messagesApi = {
  async list(channelId: string, params?: { limit?: number; offset?: number }): Promise<Message[]> {
    const response = await apiClient.get<ApiResponse<Message[]>>(`/channels/${channelId}/messages`, { params })
    return response.data.data ?? []
  },

  async create(data: MessageCreate): Promise<Message> {
    const response = await apiClient.post<ApiResponse<Message>>('/messages', data)
    return response.data.data
  },

  async update(messageId: string, data: MessageUpdate): Promise<Message> {
    const response = await apiClient.put<ApiResponse<Message>>(`/messages/${messageId}`, data)
    return response.data.data
  },

  async delete(messageId: string): Promise<void> {
    await apiClient.delete(`/messages/${messageId}`)
  },

  async getThread(parentId: string): Promise<Message[]> {
    const response = await apiClient.get<ApiResponse<Message[]>>(`/messages/${parentId}/thread`)
    return response.data.data ?? []
  },

  async markThreadAsRead(parentId: string, messageId?: string): Promise<void> {
    await apiClient.post(`/messages/${parentId}/thread/read`, { message_id: messageId })
  },

  async getThreadUnreadCount(parentId: string): Promise<number> {
    const response = await apiClient.get<ApiResponse<{ parent_id: string; unread_count: number }>>(`/messages/${parentId}/thread/unread`)
    return response.data.data?.unread_count ?? 0
  },
}

