import axios from 'axios'
import apiClient from './client'
import type { User, UserUpdate, ApiResponse } from '@/types'

function extractApiError(error: unknown, fallback: string): Error {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data as { error?: string; message?: string } | undefined
    const message = apiError?.message || apiError?.error || fallback
    return new Error(message)
  }

  if (error instanceof Error) {
    return error
  }

  return new Error(fallback)
}

export const usersApi = {
  async searchByEmail(email: string): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/users/search', {
      params: { email },
    })
    return response.data.data
  },

  async updateMe(data: UserUpdate): Promise<User> {
    try {
      const response = await apiClient.patch<ApiResponse<User>>('/auth/me', data)
      return response.data.data
    } catch (error) {
      throw extractApiError(error, 'Не удалось обновить профиль')
    }
  },

  async uploadAvatar(file: File): Promise<User> {
    const formData = new FormData()
    formData.append('avatar', file)

    try {
      const response = await apiClient.post<ApiResponse<User>>('/auth/me/avatar', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return response.data.data
    } catch (error) {
      throw extractApiError(error, 'Не удалось загрузить аватар')
    }
  },
}
