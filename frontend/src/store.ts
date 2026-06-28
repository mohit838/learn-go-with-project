import { create } from 'zustand'
import type { AuthResponse, AuthUser } from './types'

const STORAGE_KEY = 'learn_go_auth'

type AuthState = {
  user: AuthUser | null
  accessToken: string
  refreshToken: string
  hydrate: () => void
  setSession: (auth: AuthResponse) => void
  logout: () => void
}

type StoredAuth = {
  user: AuthUser
  accessToken: string
  refreshToken: string
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: '',
  refreshToken: '',
  hydrate: () => {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    try {
      const stored = JSON.parse(raw) as Partial<StoredAuth>
      if (!stored.user || !stored.accessToken || !stored.refreshToken) {
        localStorage.removeItem(STORAGE_KEY)
        return
      }
      set({
        user: stored.user,
        accessToken: stored.accessToken,
        refreshToken: stored.refreshToken,
      })
    } catch {
      localStorage.removeItem(STORAGE_KEY)
    }
  },
  setSession: (auth) => {
    const session = {
      user: auth.user,
      accessToken: auth.tokens.access_token,
      refreshToken: auth.tokens.refresh_token,
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
    set(session)
  },
  logout: () => {
    localStorage.removeItem(STORAGE_KEY)
    set({ user: null, accessToken: '', refreshToken: '' })
  },
}))

export const getAccessToken = () => useAuthStore.getState().accessToken
