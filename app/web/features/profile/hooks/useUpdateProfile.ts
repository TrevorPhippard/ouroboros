import { useMutation, useQueryClient } from "@tanstack/react-query"
import { ProfileType } from "../schemas"
import { gqlRequest } from "@/services/graphql/client"
import { UPDATE_PROFILE } from "@/lib/queries"
import { useAuthStore } from "@/store/useAuthStore"

export const useUpdateProfile = () => {
  const userId = useAuthStore((state) => state.user?.id ?? "user-1")
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (newProfile: Partial<ProfileType>) =>
      gqlRequest({
        query: UPDATE_PROFILE,
        variables: {
          userId,
          input: {
            headline: newProfile.headline,
            about: newProfile.about,
          },
        },
      }),

    onMutate: async (newProfile: Partial<ProfileType>) => {
      await queryClient.cancelQueries({ queryKey: ["profile", userId] })
      const previousProfile = queryClient.getQueryData<ProfileType>([
        "profile",
        userId,
      ])
      if (previousProfile) {
        queryClient.setQueryData<ProfileType>(["profile", userId], {
          ...previousProfile,
          ...newProfile,
        })
      }
      return { previousProfile }
    },
    onError: (err, newProfile, context) => {
      queryClient.setQueryData(["profile", userId], context?.previousProfile)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["profile", userId] })
    },
  })
}
