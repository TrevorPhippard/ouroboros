import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"
import { executeGraphQL } from "@/services/graphql/client"

// Types mirror the future GraphQL Schema / Protobufs
export interface Post {
  id: string
  author: { id: string; name: string; avatarUrl: string; headline: string }
  content: string
  likes: number
  hasLiked: boolean
  createdAt: string
}

const GET_FEED = `query GET_FEED($cursor: String) { ... }`
const LIKE_POST = `mutation LIKE_POST($postId: ID!) { ... }`

// 1. Infinite Scroll Query
export const useFeedQuery = () => {
  return useInfiniteQuery({
    queryKey: ["feed"],
    queryFn: async ({ pageParam }) =>
      executeGraphQL<{
        edges: Post[]
        pageInfo: { nextCursor: string | null }
      }>({
        query: GET_FEED,
        variables: { cursor: pageParam },
      }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pageInfo.nextCursor,
  })
}

// 2. Optimistic Like Mutation
export const useLikePost = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (postId: string) =>
      executeGraphQL({ query: LIKE_POST, variables: { postId } }),

    // ⚡ Optimistic Update Logic
    onMutate: async (postId) => {
      await queryClient.cancelQueries({ queryKey: ["feed"] })

      // Snapshot previous state for rollback
      const previousFeed = queryClient.getQueryData(["feed"])

      // Optimistically update the cache
      queryClient.setQueryData(["feed"], (oldData: any) => {
        if (!oldData) return oldData
        return {
          ...oldData,
          pages: oldData.pages.map((page: any) => ({
            ...page,
            edges: page.edges.map((post: Post) =>
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
      })

      return { previousFeed }
    },
    // Rollback on error
    onError: (err, newTodo, context) => {
      queryClient.setQueryData(["feed"], context?.previousFeed)
    },
    // Always refetch after error or success to sync with server state
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })
}
