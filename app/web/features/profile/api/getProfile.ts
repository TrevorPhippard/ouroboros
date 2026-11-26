import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "@/lib/queries"
import type { ProfileViewModel } from "../schemas"

type GetProfileResponse = {
  profile: ProfileViewModel | null
}

export const getProfile = async (
  userId: string | null
): Promise<ProfileViewModel> => {
  const response = (await gqlRequest({
    query: GET_PROFILE,
    variables: { userId },
  })) as GetProfileResponse

  const profile = response.profile

  if (!profile) {
    throw new Error("Profile not found")
  }

  return {
    id: profile.id,
    userId: profile.userId,
    username: profile.username ?? profile.userId,
    displayName: profile.displayName,
    avatarUrl: profile.avatarUrl,
    bio: profile.bio,
    headline: profile.headline,
    about: profile.about,
    experiences: profile.experiences ?? [],
    followersCount: profile.followersCount ?? 0,
    followingCount: profile.followingCount ?? 0,
    isFollowing: profile.isFollowing ?? false,
  }
}
