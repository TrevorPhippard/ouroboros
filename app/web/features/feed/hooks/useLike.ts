import { useMutation, useQueryClient } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { LIKE_POST } from "@/lib/queries"
import { PostType } from "../schemas"

export const useLikePost = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (postId: string) =>
      gqlRequest({ query: LIKE_POST, variables: { postId } }),
    onMutate: async (postId) => {
      await queryClient.cancelQueries({ queryKey: ["feed"] })
      const previousFeed = queryClient.getQueryData(["feed"])
      queryClient.setQueryData(
        ["feed"],
        (oldData: { pages: { edges: PostType[] }[] }) => {
          if (!oldData) return oldData
          return {
            ...oldData,
            pages: oldData.pages.map((page: { edges: PostType[] }) => ({
              ...page,
              edges: page.edges.map((post: PostType) =>
                post.id === postId
                  ? {
                      ...post,
                      hasLiked: !post.hasLiked,
                      likes: post.hasLiked ? post.likes - 1 : post.likes + 1,
                    }
                  : post
              ),
            })),
          }
        }
      )

      return { previousFeed }
    },
    onError: (err, newLike, context) => {
      queryClient.setQueryData(["feed"], context?.previousFeed)
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })
}
