import { gqlRequest } from "@/services/graphql/client"
import { GET_FEED } from "@/lib/queries"
import { PostType } from "../schemas"
import { feedResolvers } from "@/services/graphql/mocks/feed/resolvers"

interface FeedPageData {
  feed: {
    items: PostType[]
    pageInfo: {
      nextCursor: string | null
    }
  }
}

export const useFeedQuery = () => ({
  queryKey: ["feed"],
  // queryFn: async ({ pageParam }: { pageParam: string | null }) => {
  //   const response = await gqlRequest({
  //     query: GET_FEED,
  //     variables: { userId: "user_01", cursor: pageParam },
  //   })
  //   console.log("GraphQL Response:", response)
  //   return response
  // },
  queryFn: async ({ pageParam }: { pageParam: string | null }) => {
    const response = feedResolvers.getFeed({
      cursor: pageParam,
      limit: 2,
    })

    console.log("Mock Response:", response)
    return response
  },
  initialPageParam: null,
  getNextPageParam: (lastPage: FeedPageData) =>
    lastPage.feed.pageInfo?.nextCursor ?? undefined,
  staleTime: 1000 * 60, // 1 minute to ensure hydration sticks
})
