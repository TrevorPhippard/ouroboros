import { useMutation, useQueryClient } from "@tanstack/react-query"
import { SignInValues } from "@/features/auth/schemas"
import { useAuthStore } from "@/store/useAuthStore"
import { gqlRequest } from "@/services/graphql/client"
import { useRouter } from "next/navigation"
import { SIGN_IN } from "@/lib/queries"

type SignInResponse = {
  signIn: {
    token: string
    user: {
      id: string
      email: string
      username: string
      displayName?: string | null
    }
  }
}

export function useSignIn() {
  const queryClient = useQueryClient()
  const setAuth = useAuthStore((state) => state.setAuth)
  const router = useRouter()

  return useMutation({
    mutationFn: async (data: SignInValues) => {
      return gqlRequest({
        query: SIGN_IN,
        variables: {
          input: {
            email: data.email,
            password: data.password,
          },
        },
      })
    },
    onSuccess: (data: SignInResponse) => {
      const { user, token } = data.signIn

      setAuth(
        {
          id: user.id,
          email: user.email,
          username: user.username,
          displayName: user.displayName ?? user.username,
        },
        token
      )

      queryClient.invalidateQueries({ queryKey: ["user-session"] })

      router.push("/feed")
    },
  })
}
