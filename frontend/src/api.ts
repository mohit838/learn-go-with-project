import axios from 'axios'
import { getAccessToken } from './store'
import type {
  ApiEnvelope,
  AuthResponse,
  AuthUser,
  DashboardData,
  PaginatedResponse,
  Task,
  TaskPayload,
} from './types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:9080'

export const api = axios.create({
  baseURL: API_URL,
  timeout: 20000,
})

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

function unwrap<T>(body: ApiEnvelope<T>): T {
  if (!body.success || body.data === undefined) {
    const message =
      typeof body.error === 'string' ? body.error : body.message || 'Request failed'
    throw new Error(message)
  }
  return body.data
}

export async function login(payload: {
  tenant_slug: string
  email: string
  password: string
}) {
  const { data } = await api.post<ApiEnvelope<AuthResponse>>('/auth/login', payload)
  return unwrap(data)
}

export async function register(payload: {
  tenant_name: string
  tenant_slug: string
  username: string
  email: string
  password: string
}) {
  const { data } = await api.post<ApiEnvelope<AuthResponse>>('/auth/register', payload)
  return unwrap(data)
}

export async function listUsers(params: Record<string, unknown>) {
  const { data } = await api.get<ApiEnvelope<PaginatedResponse<AuthUser>>>(
    '/auth/users',
    { params },
  )
  return unwrap(data)
}

export async function listTasks(params: Record<string, unknown>) {
  const { data } = await api.get<ApiEnvelope<PaginatedResponse<Task>>>('/tasks', {
    params,
  })
  return unwrap(data)
}

export async function createTask(payload: TaskPayload) {
  const { data } = await api.post<ApiEnvelope<Task>>('/tasks', taskBody(payload), {
    headers: payload.image ? { 'Content-Type': 'multipart/form-data' } : undefined,
  })
  return unwrap(data)
}

export async function updateTask(id: string, payload: TaskPayload) {
  const { data } = await api.put<ApiEnvelope<Task>>(`/tasks/${id}`, taskBody(payload), {
    headers: payload.image ? { 'Content-Type': 'multipart/form-data' } : undefined,
  })
  return unwrap(data)
}

export async function markTaskInactive(id: string) {
  const { data } = await api.patch<ApiEnvelope<null>>(`/tasks/${id}/inactive`)
  return data
}

export async function deleteTask(id: string) {
  const { data } = await api.delete<ApiEnvelope<null>>(`/tasks/${id}`)
  return data
}

export async function getDashboard() {
  const query = `
    query TaskDashboard {
      dashboard {
        users {
          total
          active
          inactive
          by_role { role count }
          by_tenant { tenant_id tenant_name tenant_slug count }
        }
        tasks {
          total
          active
          inactive
          by_status { status count }
          by_priority { priority count }
          by_user { user_id total active inactive }
        }
      }
    }
  `
  const { data } = await api.post<{ data?: { dashboard: DashboardData }; errors?: Array<{ message: string }> }>(
    '/tasks/graphql',
    { query },
  )
  if (data.errors?.length) {
    throw new Error(data.errors.map((error) => error.message).join(', '))
  }
  if (!data.data?.dashboard) {
    throw new Error('Dashboard response was empty')
  }
  return data.data.dashboard
}

function taskBody(payload: TaskPayload) {
  if (!payload.image) {
    return {
      title: payload.title,
      description: payload.description,
      status: payload.status,
      priority: payload.priority,
    }
  }
  const form = new FormData()
  form.append('title', payload.title)
  form.append('description', payload.description ?? '')
  form.append('status', payload.status ?? '')
  form.append('priority', payload.priority ?? '')
  form.append('image', payload.image)
  return form
}
