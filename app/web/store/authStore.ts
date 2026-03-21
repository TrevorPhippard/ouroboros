// src/store/authStore.ts
import { create } from "zustand"

interface AuthState {
  user: { id: string; name: string; role: string } | null
  token: string | null
  setAuth: (user: AuthState["user"], token: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  setAuth: (user, token) => set({ user, token }),
  logout: () => {
    set({ user: null, token: null })
    // Note: You should also call queryClient.clear() here to wipe cached user data
  },
}))
