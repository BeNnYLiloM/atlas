import apiClient from './client'
import type { UserCreate, UserLogin, AuthResponse, User, ApiResponse, RefreshResponse } from '@/types'

export type UserStatusValue = 'online' | 'away' | 'dnd' | 'offline'

export interface AuthSession {
  id: string
  user_agent: string
  ip_address: string
  created_at: string
  last_used_at: string
  expires_at: string
  is_current: boolean
}

export const authApi = {
  async register(data: UserCreate): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data, {
      skipAuthRefresh: true,
    })
    return response.data.data
  },

  async login(data: UserLogin): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data, {
      skipAuthRefresh: true,
    })
    return response.data.data
  },

  async refresh(): Promise<RefreshResponse> {
    const response = await apiClient.post<ApiResponse<RefreshResponse>>('/auth/refresh', undefined, {
      skipAuthRefresh: true,
    })
    return response.data.data
  },

  async me(): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout', undefined, {
      skipAuthRefresh: true,
    })
  },

  async logoutAll(): Promise<void> {
    await apiClient.post('/auth/logout-all')
  },

  async listSessions(): Promise<AuthSession[]> {
    const response = await apiClient.get<ApiResponse<AuthSession[]>>('/auth/me/sessions')
    return response.data.data
  },

  async revokeSession(sessionId: string): Promise<void> {
    await apiClient.delete(`/auth/me/sessions/${sessionId}`)
  },

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await apiClient.patch('/auth/me/password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
  },

  async updateStatus(status: string, customStatus?: string | null): Promise<User> {
    const response = await apiClient.patch<ApiResponse<User>>('/auth/me/status', {
      status,
      custom_status: customStatus ?? null,
    })
    return response.data.data
  },

  async deleteAccount(password: string): Promise<void> {
    await apiClient.delete('/auth/me', { data: { password } })
  },
}
