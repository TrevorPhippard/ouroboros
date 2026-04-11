import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "@/lib/queries"

export const useUserProfile = (userId: string) => {
  return useQuery({
    queryKey: ["profile", userId],
    queryFn: () =>
      gqlRequest({
        query: GET_PROFILE,
        variables: { id: userId },
      }),
    enabled: !!userId, // Prevent fetching if ID is somehow missing
  })
}
