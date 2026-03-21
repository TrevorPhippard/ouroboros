import { useQuery } from "@tanstack/react-query"
import { executeGraphQL } from "@/services/graphql/client"

export interface UserProfile {
  id: string
  name: string
  headline: string
  avatarUrl: string
  coverUrl: string
  about: string
  followers: number
  connections: string
  location: string
}

export const useUserProfile = (userId: string) => {
  return useQuery({
    queryKey: ["profile", userId],
    queryFn: () =>
      executeGraphQL<UserProfile>({
        query: `query GET_USER_PROFILE($id: ID!) { ... }`,
        variables: { id: userId },
      }),
    enabled: !!userId, // Prevent fetching if ID is somehow missing
  })
}
