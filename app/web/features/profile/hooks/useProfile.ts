import { useQuery } from "@tanstack/react-query"
import { ProfileType } from "../schemas"
import { gqlRequest } from "@/services/graphql/client"
import { GET_UNREAD_NOTIFICATIONS } from "@/lib/queries"
import { profileResolvers } from "@/services/graphql/mocks/profile/resolvers"

export const useProfile = (userId: string) => {
  return useQuery<ProfileType>({
    queryKey: ["profile", userId],
    // queryFn: () =>
    //   gqlRequest({
    //     query: GET_UNREAD_NOTIFICATIONS,
    //     variables: { id: userId },
    //   }),

    queryFn: async () => {
      return profileResolvers.getProfile({ id: "u1" })
    },
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
  })
}
