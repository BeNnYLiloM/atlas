import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'
import type { User, UserCreate, UserLogin } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('atlas_token'))
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  async function register(data: UserCreate) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register(data)
      token.value = response.tokens.access_token
      user.value = response.user
      localStorage.setItem('atlas_token', response.tokens.access_token)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка регистрации'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function login(data: UserLogin) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login(data)
      token.value = response.tokens.access_token
      user.value = response.user
      localStorage.setItem('atlas_token', response.tokens.access_token)
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка входа'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return
    loading.value = true
    try {
      user.value = await authApi.me()
    } catch {
      logout()
    } finally {
      loading.value = false
    }
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('atlas_token')
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    register,
    login,
    fetchUser,
    logout,
  }
})

