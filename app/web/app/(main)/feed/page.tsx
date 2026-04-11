import { dehydrate, HydrationBoundary } from "@tanstack/react-query"
import { getQueryClient } from "@/lib/getQueryClient"
import { FeedList } from "@/features/feed/components/FeedList"
import { CreatePostForm } from "@/features/feed/components/CreatePostForm"

export default async function FeedPage() {
  const queryClient = getQueryClient()

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
