import { useMutation, useQueryClient } from "@tanstack/react-query"
import { SignUpValues } from "@/features/auth/schemas"

type SignUpResponse = {
  user: { id: string; email: string }
}

export function useSignUp() {
  const queryClient = useQueryClient()

  return useMutation<SignUpResponse, Error, SignUpValues>({
    mutationFn: async (data: SignUpValues) => {
      const { confirmPassword, ...apiPayload } = data
      const response = await fetch("/api/auth/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(apiPayload),
      })
      if (!response.ok) throw new Error("Failed to create account.")
      return response.json()
    },
    retry: false,
    onSuccess: (data) => {
      queryClient.setQueryData(["user-session"], data)
    },
  })
}
