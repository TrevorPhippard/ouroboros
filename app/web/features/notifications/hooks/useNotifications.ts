import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { GET_UNREAD_NOTIFICATIONS } from "@/lib/queries"

export const useNotificationsPolling = (userId: string) => {
  return useQuery({
    queryKey: ["notifications", "live"],
    queryFn: () =>
      gqlRequest({
        query: GET_UNREAD_NOTIFICATIONS,
        variables: { id: userId },
      }),
    refetchInterval: 10000, // Poll every 10 seconds to simulate event stream
    refetchIntervalInBackground: true,
  })
}
