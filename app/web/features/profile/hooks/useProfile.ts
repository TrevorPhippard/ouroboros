import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "../api/profileQueries"
import type { ProfileViewModel, UserProfileData } from "../schemas"

type GetProfileResponse = {
  user: UserProfileData | null
}

export const useProfile = (userId: string) => {
  return useQuery<ProfileViewModel, Error>({
    queryKey: ["profile", userId],
    enabled: Boolean(userId),
    queryFn: async () => {
      const response = (await gqlRequest({
        query: GET_PROFILE,
        variables: { id: userId },
      })) as GetProfileResponse

      const user = response.user
      if (!user) throw new Error("Profile not found")

      return {
        id: user.id,
        userId: user.id,
        username: user.username,
        displayName: user.displayName,
        avatarUrl: user.avatarUrl,
        bio: user.bio,
        followersCount: user.followersCount,
        followingCount: user.followingCount,
        isFollowing: user.isFollowing,
        // These fields belong to GraphQL `Profile`, but the supplied schema
        // does not expose a query that can read a Profile.
        headline: null,
        about: null,
        experiences: [],
      }
    },
    staleTime: 1000 * 60 * 5,
  })
}
