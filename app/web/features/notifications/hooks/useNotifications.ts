import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_NOTIFICATIONS } from "@/lib/queries"

export const useNotificationsPolling = (userId: string) => {
  return useQuery({
    queryKey: ["notifications", "live"],
    queryFn: () =>
      gqlRequest({
        query: GET_NOTIFICATIONS,
        variables: { userId },
      }),
    refetchInterval: 10000,
    refetchIntervalInBackground: true,
  })
}
