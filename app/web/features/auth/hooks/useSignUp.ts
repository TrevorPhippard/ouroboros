import { useMutation, useQueryClient } from "@tanstack/react-query"
import { SignUpValues } from "@/features/auth/schemas"
import { gqlRequest } from "@/services/graphql/client"
import { SIGN_UP } from "@/lib/queries"
import { useRouter } from "next/navigation"

export type SignUpResponse = {
  signUp: {
    token: string
    user: {
      id: string
      email: string
      username: string
      displayName: string
    }
  }
}

export function useSignUp() {
  const queryClient = useQueryClient()
  const router = useRouter()

  return useMutation<SignUpResponse, Error, SignUpValues>({
    mutationFn: async (data) => {
      const { confirmPassword: _confirmPassword, name, ...rest } = data

      return gqlRequest({
        query: SIGN_UP,
        variables: {
          input: {
            ...rest,
            displayName: name,
          },
        },
      })
    },

    retry: false,

    onSuccess: (data) => {
      // Registration succeeded. Redirect to login.
      queryClient.invalidateQueries({ queryKey: ["user-session"] })
      router.push("/login")
    },
  })
}
