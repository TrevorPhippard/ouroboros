import { dehydrate, HydrationBoundary } from "@tanstack/react-query"
import { getQueryClient } from "@/lib/getQueryClient"
import { executeGraphQL } from "@/services/graphql/client"
import { FeedList } from "@/features/feed/components/FeedList"
import { CreatePostForm } from "@/features/feed/components/CreatePostForm"

// The GraphQL query string (matches what the client hook uses)
const GET_FEED = `query GET_FEED($cursor: String) { ... }`

export default async function FeedPage() {
  const queryClient = getQueryClient()

  // Prefetch the infinite query's first page on the server
  await queryClient.prefetchInfiniteQuery({
    queryKey: ["feed"],
    queryFn: async () =>
      executeGraphQL({
        query: GET_FEED,
        variables: { cursor: null },
      }),
    initialPageParam: null,
  })

  return (
    <main className="mx-auto max-w-2xl space-y-6 py-6">
      <CreatePostForm />

      {/* HydrationBoundary injects the server-fetched data into TanStack Query's
        client-side cache before the FeedList component even mounts.
      */}
      <HydrationBoundary state={dehydrate(queryClient)}>
        <FeedList />
      </HydrationBoundary>
    </main>
  )
}
