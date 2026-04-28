import { gqlRequest } from "@/services/graphql/client"
import { GET_FEED } from "@/lib/queries"
import { PostType } from "../schemas"
import { useAuthStore } from "@/store/useAuthStore"

interface FeedPageData {
  feed: {
    items: PostType[]
    nextCursor?: string | null
  }
}

export const useFeedQuery = () => {
  const userId = useAuthStore((state) => state.user?.id ?? "user-1")

  return {
    queryKey: ["feed", userId],
    queryFn: async ({ pageParam }: { pageParam: string | null }) => {
      const response = await gqlRequest({
        query: GET_FEED,
        variables: { userId, cursor: pageParam },
      })

      return {
        feed: {
          items:
            response.feed?.items?.flatMap((item: any) => {
              if (!item.post?.author) return []
              return [
                {
                  id: item.post.id ?? item.postId,
                  content: item.post.content,
                  createdAt: item.post.createdAt ?? item.cursor,
                  likes: 0,
                  hasLiked: false,
                  author: {
                    id: item.post.author.id,
                    username: item.post.author.username,
                    name:
                      item.post.author.displayName ?? item.post.author.username,
                    avatarUrl: item.post.author.avatarUrl,
                    headline: item.post.author.bio ?? undefined,
                  },
                } satisfies PostType,
              ]
            }) ?? [],
          nextCursor: response.feed?.nextCursor ?? null,
        },
      } satisfies FeedPageData
    },
    initialPageParam: null,
    getNextPageParam: (lastPage: FeedPageData) =>
      lastPage.feed.nextCursor ?? undefined,
    staleTime: 1000 * 60,
  }
}
