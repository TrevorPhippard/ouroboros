import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_PROFILE } from "@/lib/queries"
import type { ProfileViewModel } from "../schemas"
import { getProfile } from "../api/getProfile"

export const useProfile = (userId: string) => {
  return useQuery<ProfileViewModel, Error>({
    queryKey: ["profile", userId],
    enabled: Boolean(userId),
    queryFn: () => getProfile(userId),
    staleTime: 0,
  })
}
