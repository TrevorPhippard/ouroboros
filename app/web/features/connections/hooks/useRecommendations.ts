import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_RECOMMENDATIONS } from "@/lib/queries"

export const useRecommendations = () => {
  return useQuery({
    queryKey: ["recommendations"],
    queryFn: async () => {
      const response = await gqlRequest({ query: GET_RECOMMENDATIONS })
      return (
        response.recommendations?.map((user: any) => ({
          id: user.id,
          username: user.username,
          name: user.displayName ?? user.username,
          headline: user.bio ?? "",
          avatarUrl: user.avatarUrl ?? "",
        })) ?? []
      )
    },
  })
}
