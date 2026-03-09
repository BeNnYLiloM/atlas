import apiClient from './client'
import type { UserCreate, UserLogin, AuthResponse, User, ApiResponse, RefreshResponse } from '@/types'

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
}
