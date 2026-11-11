import { useMutation, useQueryClient } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { FOLLOW_USER, UNFOLLOW_USER } from "../api/profileQueries"
import type { ProfileViewModel } from "../schemas"

type ConnectResponse = { success: boolean }
type FollowResponse = { followUser: ConnectResponse }
type UnfollowResponse = { unfollowUser: ConnectResponse }
type FollowContext = { previousProfile?: ProfileViewModel }

export const useToggleFollow = (targetUserId: string) => {
  const queryClient = useQueryClient()
  const queryKey = ["profile", targetUserId]

  return useMutation<void, Error, boolean, FollowContext>({
    mutationFn: async (shouldFollow) => {
      const response = shouldFollow
        ? ((await gqlRequest({
            query: FOLLOW_USER,
            variables: { userId: targetUserId },
          })) as FollowResponse)
        : ((await gqlRequest({
            query: UNFOLLOW_USER,
            variables: { userId: targetUserId },
          })) as UnfollowResponse)

      const result = shouldFollow
        ? (response as FollowResponse).followUser
        : (response as UnfollowResponse).unfollowUser

      if (!result?.success) {
        throw new Error(shouldFollow ? "Could not follow this user." : "Could not unfollow this user.")
      }
    },

    onMutate: async (shouldFollow) => {
      await queryClient.cancelQueries({ queryKey })
      const previousProfile = queryClient.getQueryData<ProfileViewModel>(queryKey)

      if (previousProfile) {
        queryClient.setQueryData<ProfileViewModel>(queryKey, {
          ...previousProfile,
          isFollowing: shouldFollow,
          followersCount: Math.max(
            0,
            previousProfile.followersCount + (shouldFollow ? 1 : -1),
          ),
        })
      }

      return { previousProfile }
    },

    onError: (_error, _shouldFollow, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(queryKey, context.previousProfile)
      }
    },
  })
}
