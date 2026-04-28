import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "@/lib/queries"

export const useUserProfile = (userId: string) => {
  return useQuery({
    queryKey: ["profile", userId],
    queryFn: async () => {
      const response = await gqlRequest({
        query: GET_PROFILE,
        variables: { userId },
      })
      const user = response.user
      return {
        id: user.id,
        name: user.displayName ?? user.username,
        username: user.username,
        avatarUrl: user.avatarUrl ?? "",
        headline: user.bio ?? "",
        about: user.bio ?? "",
        followers: user.followersCount ?? 0,
        connections: String(user.followingCount ?? 0),
      }
    },
    enabled: !!userId,
  })
}
