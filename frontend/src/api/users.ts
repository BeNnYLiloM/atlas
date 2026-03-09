import apiClient from './client'
import type { User, ApiResponse } from '@/types'

export const usersApi = {
  async searchByEmail(email: string): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/users/search', {
      params: { email },
    })
    return response.data.data
  },
}

