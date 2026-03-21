import { create } from "zustand"
import { persist, createJSONStorage } from "zustand/middleware"

interface User {
  id: string
  email: string
  name: string
  avatarUrl: string
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
      setAuth: (user, token) => set({ user, token, isAuthenticated: true }),
      logout: () => {
        set({ user: null, token: null, isAuthenticated: false })
        // Optional: clear any sensitive caches
        window.location.href = "/login"
      },
    }),
    {
      name: "linkedin-auth-storage", // Key in localStorage
      storage: createJSONStorage(() => localStorage),
    }
  )
)
