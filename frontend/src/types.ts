export type ApiEnvelope<T> = {
  success: boolean
  message: string
  data?: T
  error?: unknown
}

export type PaginationMeta = {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export type PaginatedResponse<T> = {
  items: T[]
  meta: PaginationMeta
}

export type AuthUser = {
  id: string
  tenant_id: string
  tenant_name: string
  tenant_slug: string
  role: string
  username: string
  email: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type AuthTokens = {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export type AuthResponse = {
  user: AuthUser
  tokens: AuthTokens
}

export type Task = {
  id: string
  tenant_id: string
  tenant_slug: string
  user_id: string
  title: string
  description?: string
  status: 'todo' | 'in_progress' | 'done'
  priority: 'low' | 'normal' | 'high'
  image_url?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export type TaskPayload = {
  title: string
  description?: string
  status?: string
  priority?: string
  image?: File
}

export type DashboardData = {
  users: {
    total: number
    active: number
    inactive: number
    by_role: Array<{ role: string; count: number }>
    by_tenant: Array<{
      tenant_id: string
      tenant_name: string
      tenant_slug: string
      count: number
    }>
  }
  tasks: {
    total: number
    active: number
    inactive: number
    by_status: Array<{ status: string; count: number }>
    by_priority: Array<{ priority: string; count: number }>
    by_user: Array<{
      user_id: string
      total: number
      active: number
      inactive: number
    }>
  }
}
