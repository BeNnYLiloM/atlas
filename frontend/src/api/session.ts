import { ref } from 'vue'
import axios from 'axios'
import type { ApiResponse, AuthTokens, RefreshResponse } from '@/types'

const configuredApiUrl = import.meta.env.VITE_API_URL?.trim()
const apiBaseUrl = configuredApiUrl || '/api/v1'
const refreshSkewSeconds = 30

const refreshClient = axios.create({
  baseURL: apiBaseUrl,
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
    'X-Requested-With': 'XMLHttpRequest',
  },
})

const accessToken = ref<string | null>(null)
const accessTokenExpiresAt = ref<number | null>(null)
let refreshPromise: Promise<string | null> | null = null

export function getApiBaseUrl(): string {
  return apiBaseUrl
}

export function getAccessToken(): string | null {
  if (!accessToken.value || isTokenExpired()) {
    return null
  }
  return accessToken.value
}

export function setAccessToken(token: string | null, expiresAt: number | null = null) {
  accessToken.value = token
  accessTokenExpiresAt.value = expiresAt
}

export function clearAccessToken() {
  accessToken.value = null
  accessTokenExpiresAt.value = null
}

export async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) {
    return refreshPromise
  }

  refreshPromise = refreshClient
    .post<ApiResponse<RefreshResponse>>('/auth/refresh')
    .then((response) => {
      applyAuthTokens(response.data.data.tokens)
      return response.data.data.tokens.access_token
    })
    .catch(() => {
      clearAccessToken()
      return null
    })
    .finally(() => {
      refreshPromise = null
    })

  return refreshPromise
}

export async function ensureAccessToken(): Promise<string | null> {
  if (accessToken.value && !isTokenExpired()) {
    return accessToken.value
  }
  return refreshAccessToken()
}

export function applyAuthTokens(tokens: AuthTokens) {
  setAccessToken(tokens.access_token, tokens.expires_at)
}

function isTokenExpired(): boolean {
  if (!accessToken.value || !accessTokenExpiresAt.value) {
    return true
  }
  const now = Math.floor(Date.now() / 1000)
  return accessTokenExpiresAt.value <= now+refreshSkewSeconds
}
