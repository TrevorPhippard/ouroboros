import { useQuery } from "@tanstack/react-query"
import { executeGraphQL } from "@/services/graphql/client"

export const useNotificationsPolling = () => {
  return useQuery({
    queryKey: ["notifications", "live"],
    queryFn: () =>
      executeGraphQL({ query: `query GET_UNREAD_NOTIFICATIONS { ... }` }),
    refetchInterval: 10000, // Poll every 10 seconds to simulate event stream
    refetchIntervalInBackground: true,
  })
}
