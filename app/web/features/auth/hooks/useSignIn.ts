import { useMutation, useQueryClient } from "@tanstack/react-query"
import { SignInValues } from "@/features/auth/schemas"
import { useAuthStore } from "@/store/useAuthStore"
import { gqlRequest } from "@/services/graphql/client"
import { useRouter } from "next/navigation"
import { SIGN_IN } from "@/lib/queries"

export function useSignIn() {
  const queryClient = useQueryClient()
  const setAuth = useAuthStore((state) => state.setAuth)
  const router = useRouter()

  return useMutation({
    mutationFn: async (data: SignInValues) => {
      console.log("Attempting to sign in with:", data)
      return gqlRequest({
        query: SIGN_IN,
        variables: {
          input: data,
        },
      })
    },
    onSuccess: (data: { signIn: { user: any; token: string } }) => {
      const { user, token } = data.signIn

      setAuth(user, token)

      queryClient.invalidateQueries({ queryKey: ["user-session"] })

      router.push("/feed")
    },
  })
}
