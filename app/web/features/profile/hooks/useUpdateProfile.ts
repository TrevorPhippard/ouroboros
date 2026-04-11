import { useMutation, useQueryClient } from "@tanstack/react-query"
import { ProfileType } from "../schemas"
import { updateProfileApi } from "../api/profileApi" // Mock fetcher

export const useUpdateProfile = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (newProfile: Partial<ProfileType>) =>
      updateProfileApi("currentId", newProfile),

    onMutate: async (newProfile: Partial<ProfileType>) => {
      await queryClient.cancelQueries({ queryKey: ["profile"] })
      const previousProfile = queryClient.getQueryData<ProfileType>(["profile"])
      if (previousProfile) {
        queryClient.setQueryData<ProfileType>(["profile"], {
          ...previousProfile,
          ...newProfile,
        })
      }
      return { previousProfile }
    },
    onError: (err, newProfile, context) => {
      queryClient.setQueryData(["profile"], context?.previousProfile)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] })
    },
  })
}
