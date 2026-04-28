import { useMutation, useQueryClient } from "@tanstack/react-query"
import { SignUpValues } from "@/features/auth/schemas"
import { gqlRequest } from "@/services/graphql/client"
import { SIGN_UP } from "@/lib/queries"
import { useRouter } from "next/navigation"

type SignUpResponse = {
  signUp: {
    id: string
    displayName: string
    email: string
  }
}

export function useSignUp() {
  const queryClient = useQueryClient()
  const router = useRouter()

  return useMutation<SignUpResponse, Error, SignUpValues>({
    mutationFn: async (data: SignUpValues) => {
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
      queryClient.setQueryData(["user-session"], data)
      router.push("/login")
    },
  })
}
