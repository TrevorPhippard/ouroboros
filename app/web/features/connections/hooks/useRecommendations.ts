import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_RECOMMENDATIONS } from "@/lib/queries"

export const useRecommendations = () => {
  return useQuery({
    queryKey: ["recommendations"],
    queryFn: () => gqlRequest({ query: GET_RECOMMENDATIONS }),
  })
}
