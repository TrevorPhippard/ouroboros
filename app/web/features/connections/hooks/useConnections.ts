import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { executeGraphQL } from "@/services/graphql/client"

export const useRecommendations = () => {
  return useQuery({
    queryKey: ["recommendations"],
    queryFn: () =>
      executeGraphQL<any>({ query: `query GET_RECOMMENDATIONS { ... }` }),
  })
}

export const useSendConnectionRequest = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (userId: string) =>
      executeGraphQL({
        query: `mutation SEND_CONNECT($userId: ID!) { ... }`,
        variables: { userId },
      }),

    onMutate: async (userId) => {
      // 1. Cancel outgoing fetches
      await queryClient.cancelQueries({ queryKey: ["recommendations"] })

      // 2. Snapshot current state
      const previousData = queryClient.getQueryData(["recommendations"])

      // 3. Optimistically remove the user from the list
      queryClient.setQueryData(["recommendations"], (old: any) => ({
        ...old,
        edges: old.edges.filter((user: any) => user.id !== userId),
      }))

      return { previousData }
    },
    // 4. If the microservice fails, roll back to previous state
    onError: (err, userId, context) => {
      queryClient.setQueryData(["recommendations"], context?.previousData)
      alert("Failed to send request. Please try again.")
    },
    // 5. Refetch to ensure we are in sync with the server
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["recommendations"] })
    },
  })
}
