import { useQuery } from "@tanstack/react-query"
import { ProfileType } from "../schemas"
import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "@/lib/queries"

export const useProfile = (userId: string) => {
  return useQuery<ProfileType>({
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
        avatarUrl: user.avatarUrl ?? undefined,
        headline: user.bio ?? "Add a short professional headline",
        about: user.bio ?? "",
        followersCount: user.followersCount ?? 0,
        followingCount: user.followingCount ?? 0,
        experiences: [],
      }
    },
    staleTime: 1000 * 60 * 5,
  })
}
