import apiClient from './client'

export type TaskStatus = 'todo' | 'in_progress' | 'done' | 'cancelled'
export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent'

export interface Task {
  id: string
  message_id: string | null
  workspace_id: string
  project_id: string | null
  title: string
  description: string | null
  status: TaskStatus
  priority: TaskPriority
  assignee_id: string | null
  reporter_id: string | null
  due_date: string | null
  created_at: string
  updated_at: string
}

export interface TaskCreate {
  message_id?: string
  workspace_id: string
  project_id?: string
  title: string
  description?: string
  priority?: TaskPriority
  assignee_id?: string
  due_date?: string
}

export interface TaskUpdate {
  status?: TaskStatus
  priority?: TaskPriority
  assignee_id?: string | null
  title?: string
  due_date?: string | null
}

export const tasksApi = {
  async create(data: TaskCreate): Promise<Task> {
    const res = await apiClient.post<{ data: Task }>('/tasks', data)
    return res.data.data
  },

  async list(workspaceId: string, options?: { projectId?: string; status?: string }): Promise<Task[]> {
    const params = new URLSearchParams({ workspace_id: workspaceId })
    if (options?.projectId) params.append('project_id', options.projectId)
    if (options?.status) params.append('status', options.status)
    const res = await apiClient.get<{ data: Task[] }>(`/tasks?${params}`)
    return res.data.data ?? []
  },

  async update(id: string, update: TaskUpdate): Promise<void> {
    await apiClient.patch(`/tasks/${id}`, update)
  },

  async delete(id: string): Promise<void> {
    await apiClient.delete(`/tasks/${id}`)
  },
}

export const TASK_STATUS_LABELS: Record<TaskStatus, string> = {
  todo: 'К выполнению',
  in_progress: 'В работе',
  done: 'Готово',
  cancelled: 'Отменено',
}

export const TASK_PRIORITY_LABELS: Record<TaskPriority, string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
  urgent: 'Срочно',
}

export const TASK_PRIORITY_COLORS: Record<TaskPriority, string> = {
  low: 'text-dark-400',
  medium: 'text-blue-400',
  high: 'text-amber-400',
  urgent: 'text-red-400',
}

export const KANBAN_COLUMNS: TaskStatus[] = ['todo', 'in_progress', 'done']
