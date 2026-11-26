import { useQuery } from "@tanstack/react-query"
import type { ProfileViewModel } from "../schemas"
import { getProfile } from "../api/getProfile"

export const useProfile = (userId: string | null) => {
  return useQuery<ProfileViewModel, Error>({
    queryKey: ["profile", userId],
    enabled: Boolean(userId),
    queryFn: () => getProfile(userId),
    staleTime: 0,
  })
}
