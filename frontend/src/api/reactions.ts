import apiClient from './client'

export interface ReactionGroup {
  emoji: string
  count: number
  user_ids: string[]
  mine: boolean
}

export const reactionsApi = {
  async add(messageId: string, emoji: string, workspaceId: string): Promise<void> {
    await apiClient.post(`/messages/${messageId}/reactions?workspace_id=${workspaceId}`, { emoji })
  },

  async remove(messageId: string, emoji: string, workspaceId: string): Promise<void> {
    await apiClient.delete(`/messages/${messageId}/reactions/${encodeURIComponent(emoji)}?workspace_id=${workspaceId}`)
  },

  async getGrouped(messageId: string): Promise<ReactionGroup[]> {
    const res = await apiClient.get<{ data: ReactionGroup[] }>(`/messages/${messageId}/reactions`)
    return res.data.data ?? []
  },
}

// Популярные эмодзи для быстрого доступа
export const QUICK_EMOJIS = ['👍', '❤️', '😂', '🎉', '😮', '😢', '😡', '👀', '🔥', '✅']
