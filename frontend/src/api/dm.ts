import apiClient from './client'
import type { ApiResponse } from '@/types'

export interface DMPeer {
  userId: string
  displayName: string
  avatarUrl: string | null
  status: string
}

export interface DMChannel {
  channelId: string
  workspaceId: string
  peer: DMPeer
  unreadCount: number
  lastMessageAt: string | null
  createdAt: string
}

export interface OpenDMRequest {
  workspace_id: string
  target_user_id: string
}

// Минимальное представление канала, возвращаемое при открытии DM
export interface DMOpenResult {
  id: string
  workspace_id: string
  type: 'dm'
  created_at: string
}

function mapPeer(raw: { user_id: string; display_name: string; avatar_url: string | null; status: string }): DMPeer {
  return {
    userId: raw.user_id,
    displayName: raw.display_name,
    avatarUrl: raw.avatar_url,
    status: raw.status,
  }
}

function mapDMChannel(raw: {
  channel_id: string
  workspace_id: string
  peer: { user_id: string; display_name: string; avatar_url: string | null; status: string }
  unread_count: number
  last_message_at: string | null
  created_at: string
}): DMChannel {
  return {
    channelId: raw.channel_id,
    workspaceId: raw.workspace_id,
    peer: mapPeer(raw.peer),
    unreadCount: raw.unread_count ?? 0,
    lastMessageAt: raw.last_message_at ?? null,
    createdAt: raw.created_at,
  }
}

export const dmApi = {
  async open(workspaceId: string, targetUserId: string): Promise<DMOpenResult> {
    const payload: OpenDMRequest = { workspace_id: workspaceId, target_user_id: targetUserId }
    const res = await apiClient.post<ApiResponse<DMOpenResult>>('/dm', payload)
    return res.data.data
  },

  async list(workspaceId: string): Promise<DMChannel[]> {
    const res = await apiClient.get<ApiResponse<unknown[]>>('/dm', { params: { workspace_id: workspaceId } })
    const raw = res.data.data ?? []
    return raw.map((item) =>
      mapDMChannel(
        item as {
          channel_id: string
          workspace_id: string
          peer: { user_id: string; display_name: string; avatar_url: string | null; status: string }
          unread_count: number
          last_message_at: string | null
          created_at: string
        },
      ),
    )
  },
}
