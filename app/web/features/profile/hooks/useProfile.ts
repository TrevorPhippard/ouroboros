// src/features/profile/hooks/useProfile.ts
import { useQuery } from "@tanstack/react-query"
import { getProfileApi } from "../api/profileApi" // Mock fetcher
import { Profile } from "../schemas"

export const useProfile = (userId: string) => {
  return useQuery<Profile>({
    queryKey: ["profile", userId],
    queryFn: () => getProfileApi(userId),
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
  })
}
