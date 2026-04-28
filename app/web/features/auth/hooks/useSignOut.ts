import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "@/store/useAuthStore"
import { gqlRequest } from "@/services/graphql/client"
import { SIGN_OUT } from "@/lib/queries"

export function useSignOut() {
  const queryClient = useQueryClient()
  const { logout } = useAuthStore()

  return useMutation({
    mutationFn: async () => {
      return gqlRequest({ query: SIGN_OUT })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user-session"] })
      logout()
    },
  })
}
