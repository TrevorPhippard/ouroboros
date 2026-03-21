import { graphqlClient } from "@/lib/graphqlClient"
import { CREATE_POST } from "@/lib/queries"
import { useQueryClient, useMutation } from "@tanstack/react-query"

export function useCreatePost() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: { content: string }) => {
      const data = await graphqlClient.request(CREATE_POST, { input })
      return data.createPost
    },

    onMutate: async (newPost) => {
      await queryClient.cancelQueries({ queryKey: ["feed"] })

      const previousFeed = queryClient.getQueryData(["feed"])

      queryClient.setQueryData(["feed"], (old: any) => [
        {
          id: "temp-id",
          content: newPost.content,
          createdAt: new Date().toISOString(),
        },
        ...(old || []),
      ])

      return { previousFeed }
    },

    onError: (_err, _newPost, context) => {
      queryClient.setQueryData(["feed"], context?.previousFeed)
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })
}
