import { useMutation, useQueryClient } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { UPDATE_PROFILE } from "@/lib/queries"
import { useAuthStore } from "@/store/useAuthStore"
import type {
  ProfileType,
  ProfileUpdateInput,
  ProfileViewModel,
} from "../schemas"

type UpdateProfileResponse = {
  updateProfile: ProfileType
}

const toGraphQLInput = (updates: ProfileUpdateInput): ProfileUpdateInput => {
  const input: ProfileUpdateInput = {}
  const fields: (keyof ProfileUpdateInput)[] = [
    "headline",
    "about",
    "displayName",
    "avatarUrl",
    "bio",
  ]

  for (const field of fields) {
    const value = updates[field]
    if (value !== undefined) input[field] = value
  }

  return input
}

export const useUpdateProfile = (profileUserId?: string) => {
  const authenticatedUserId = useAuthStore((state) => state.user?.id)
  const userId = profileUserId ?? authenticatedUserId
  const queryClient = useQueryClient()
  const queryKey = ["profile", userId]

  return useMutation<
    ProfileType,
    Error,
    ProfileUpdateInput,
    { previousProfile?: ProfileViewModel }
  >({
    mutationFn: async (updates) => {
      if (!userId) throw new Error("You must be signed in to update a profile.")

      const input = toGraphQLInput(updates)
      if (Object.keys(input).length === 0) {
        throw new Error("There are no profile changes to save.")
      }

      const response = (await gqlRequest({
        query: UPDATE_PROFILE,
        variables: { userId, input },
      })) as UpdateProfileResponse

      if (!response?.updateProfile) {
        throw new Error("The profile update did not return a profile.")
      }

      return response.updateProfile
    },

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey })
      const previousProfile =
        queryClient.getQueryData<ProfileViewModel>(queryKey)

      if (previousProfile) {
        queryClient.setQueryData<ProfileViewModel>(queryKey, {
          ...previousProfile,
          ...updates,
        })
      }

      return { previousProfile }
    },

    onSuccess: (updatedProfile) => {
      queryClient.setQueryData<ProfileViewModel>(
        queryKey,
        (previousProfile) => {
          if (!previousProfile) return previousProfile

          // Only apply fields actually returned by the API.
          const definedUpdates = Object.fromEntries(
            Object.entries(updatedProfile).filter(
              ([, value]) => value !== undefined
            )
          )

          return {
            ...previousProfile,
            ...definedUpdates,
            id: previousProfile.id,
            userId: previousProfile.userId,
            username: previousProfile.username,
            followersCount: previousProfile.followersCount,
            followingCount: previousProfile.followingCount,
            isFollowing: previousProfile.isFollowing,
          }
        }
      )
    },

    onError: (_error, _updates, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(queryKey, context.previousProfile)
      }
    },
  })
}
