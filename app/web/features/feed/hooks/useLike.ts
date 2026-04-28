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
      const previousFeed = queryClient.getQueriesData({ queryKey: ["feed"] })
      const firstFeedKey = previousFeed[0]?.[0]
      if (!firstFeedKey) {
        return { previousFeed }
      }

      queryClient.setQueryData(
        firstFeedKey,
        (oldData: { pages: { feed: { items: PostType[] } }[] }) => {
          if (!oldData) return oldData
          return {
            ...oldData,
            pages: oldData.pages.map((page: { feed: { items: PostType[] } }) => ({
              ...page,
              feed: {
                ...page.feed,
                items: page.feed.items.map((post: PostType) =>
                  post.id === postId
                    ? {
                        ...post,
                        hasLiked: !post.hasLiked,
                        likes: post.hasLiked ? post.likes - 1 : post.likes + 1,
                      }
                    : post
                ),
              },
            })),
          }
        }
      )

      return { previousFeed }
    },
    onError: (err, newLike, context) => {
      context?.previousFeed?.forEach(([key, data]) => {
        queryClient.setQueryData(key, data)
      })
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })
}
