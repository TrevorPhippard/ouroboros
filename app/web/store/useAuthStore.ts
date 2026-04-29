import { create } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"

interface User {
  id: string
  email: string
  username?: string
  displayName: string
  avatarUrl?: string
  bio?: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  setAuth: (user: User, token: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      setAuth: (user, token) => {
        if (typeof window !== "undefined") {
          localStorage.setItem("token", token)
        }
        set({ user, token, isAuthenticated: true })
      },
      logout: () => {
        if (typeof window !== "undefined") {
          localStorage.removeItem("token")
        }
        set({ user: null, token: null, isAuthenticated: false })
        window.location.href = "/login"
      },
    }),
    {
      name: "linkedin-auth-storage", // Key in localStorage
      storage: createJSONStorage(() => localStorage),
    }
  )
)
