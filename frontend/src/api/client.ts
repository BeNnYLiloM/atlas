import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

const configuredApiUrl = import.meta.env.VITE_API_URL?.trim()

const apiClient: AxiosInstance = axios.create({
  baseURL: configuredApiUrl || '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('atlas_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('atlas_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient
