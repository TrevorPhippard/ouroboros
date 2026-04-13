import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "@/store/useAuthStore"

export function useSignOut() {
  const queryClient = useQueryClient()
  const { logout } = useAuthStore()

  return useMutation({
    mutationFn: async () => {
      const response = await fetch("/api/auth/signout", { method: "POST" })
      if (!response.ok) throw new Error("Failed to sign out.")
      return response.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-session"] })
      logout()
    },
  })
}
