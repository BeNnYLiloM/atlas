import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { authApi, usersApi } from '@/api'
import { applyAuthTokens, clearAccessToken, getAccessToken, refreshAccessToken } from '@/api/session'
import type { User, UserCreate, UserLogin, UserUpdate } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const initialized = ref(false)

  const token = computed(() => getAccessToken())
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  async function register(data: UserCreate) {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register(data)
      applyAuthTokens(response.tokens)
      user.value = response.user
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
      applyAuthTokens(response.tokens)
      user.value = response.user
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка входа'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function initialize() {
    if (initialized.value) {
      return
    }

    loading.value = true
    error.value = null
    try {
      const nextToken = await refreshAccessToken()
      if (!nextToken) {
        clearState()
        return
      }
      user.value = await authApi.me()
    } catch {
      clearState()
    } finally {
      initialized.value = true
      loading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return
    loading.value = true
    try {
      user.value = await authApi.me()
    } catch {
      clearState()
      throw new Error('failed to fetch user')
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data: UserUpdate) {
    if (!user.value) {
      throw new Error('user is not loaded')
    }

    loading.value = true
    error.value = null
    try {
      user.value = await usersApi.updateMe(data)
      return user.value
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка обновления профиля'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function uploadAvatar(file: File) {
    if (!user.value) {
      throw new Error('user is not loaded')
    }

    loading.value = true
    error.value = null
    try {
      user.value = await usersApi.uploadAvatar(file)
      return user.value
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Ошибка загрузки аватара'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch {
      // Локально завершаем сессию даже если запрос logout не дошел.
    } finally {
      clearState()
    }
  }

  async function logoutAll() {
    try {
      await authApi.logoutAll()
    } finally {
      clearState()
    }
  }

  function clearState() {
    user.value = null
    clearAccessToken()
  }

  return {
    user,
    token,
    loading,
    error,
    initialized,
    isAuthenticated,
    register,
    login,
    initialize,
    fetchUser,
    updateProfile,
    uploadAvatar,
    logout,
    logoutAll,
  }
})
