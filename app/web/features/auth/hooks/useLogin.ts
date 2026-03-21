import { useMutation } from "@tanstack/react-query"
import { useAuthStore } from "@/store/useAuthStore"
import { executeGraphQL } from "@/services/graphql/client"
import { useRouter } from "next/navigation"

export const useLogin = () => {
  const setAuth = useAuthStore((state) => state.setAuth)
  const router = useRouter()

  return useMutation({
    mutationFn: async (credentials: { email: string }) => {
      // Simulate calling the Auth Microservice
      return executeGraphQL<{ login: { user: any; token: string } }>({
        query: `mutation LOGIN($email: String!) { ... }`,
        variables: credentials,
      })
    },
    onSuccess: (data) => {
      // Store the "JWT" in our global state
      setAuth(data.login.user, data.login.token)
      router.push("/feed")
    },
  })
}
