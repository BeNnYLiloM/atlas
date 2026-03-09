import apiClient from './client'
import type { UserCreate, UserLogin, AuthResponse, User, ApiResponse } from '@/types'

export const authApi = {
  async register(data: UserCreate): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data)
    return response.data.data
  },

  async login(data: UserLogin): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data)
    return response.data.data
  },

  async me(): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  },

  async logout(): Promise<void> {
    // Бэкенд не имеет эндпоинта logout — просто удаляем токен локально
  },
}

