import { useQuery } from "@tanstack/react-query"
import { ProfileType } from "../schemas"
import { gqlRequest } from "@/services/graphql/client"
import { GET_UNREAD_NOTIFICATIONS } from "@/lib/queries"

export const useProfile = (userId: string) => {
  return useQuery<ProfileType>({
    queryKey: ["profile", userId],
    queryFn: () =>
      gqlRequest<ProfileType>({
        query: GET_UNREAD_NOTIFICATIONS,
        variables: { id: userId },
      }),
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
  })
}
