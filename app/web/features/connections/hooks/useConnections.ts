import { useMutation, useQueryClient } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { SEND_CONNECT } from "@/lib/queries"
import { Recommendation, EdgeUser } from "../schemas"

export const useSendConnectionRequest = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (userId: string) =>
      gqlRequest({
        query: SEND_CONNECT,
        variables: { userId },
      }),
    onMutate: async (userId) => {
      await queryClient.cancelQueries({ queryKey: ["recommendations"] })
      const previousData = queryClient.getQueryData(["recommendations"])
      queryClient.setQueryData(["recommendations"], (old: Recommendation) => ({
        ...old,
        edges: old.edges.filter((user: EdgeUser) => user.id !== userId),
      }))

      return { previousData }
    },
    onError: (err, userId, context) => {
      queryClient.setQueryData(["recommendations"], context?.previousData)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["recommendations"] })
    },
  })
}
