import apiClient from './client'

export interface UploadedFile {
  id: string
  message_id: string | null
  user_id: string
  filename: string
  original_name: string
  mime_type: string
  size_bytes: number
  url: string
  created_at: string
}

export const filesApi = {
  async upload(file: File): Promise<UploadedFile> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await apiClient.post<{ data: UploadedFile }>('/files/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data.data
  },

  async getById(id: string): Promise<UploadedFile> {
    const response = await apiClient.get<{ data: UploadedFile }>(`/files/${id}`)
    return response.data.data
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/files/${id}`)
  },
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function isImageFile(mimeType: string): boolean {
  return mimeType.startsWith('image/')
}

export const FILE_LIMITS = {
  free: {
    maxSizeBytes: 10 * 1024 * 1024, // 10 MB
    allowedTypes: ['image/', 'application/pdf', 'text/'],
  },
  pro: {
    maxSizeBytes: 100 * 1024 * 1024, // 100 MB
    allowedTypes: ['*'],
  },
}
