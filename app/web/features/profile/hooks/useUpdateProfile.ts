import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Profile } from "../schemas"
import { updateProfileApi } from "../api/profileApi" // Mock fetcher

export const useUpdateProfile = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (newProfile: Partial<Profile>) =>
      updateProfileApi("currentId", newProfile),

    // 1. Optimistic Update
    onMutate: async (newProfile) => {
      await queryClient.cancelQueries({ queryKey: ["profile"] })

      const previousProfile = queryClient.getQueryData<Profile>(["profile"])

      // Optimistically update the cache
      if (previousProfile) {
        queryClient.setQueryData<Profile>(["profile"], {
          ...previousProfile,
          ...newProfile,
        })
      }
      return { previousProfile }
    },

    // 2. Rollback on Error
    onError: (err, newProfile, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(["profile"], context.previousProfile)
      }
      // Trigger a toast notification here: toast.error("Failed to update profile")
    },

    // 3. Sync with Server on Settlement
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] })
    },
  })
}
