"use client"

import { useFeedQuery } from "../hooks/useFeed"
import { PostCard } from "./PostCard" // Assume this is a standard UI component

export const FeedList = () => {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, status } =
    useFeedQuery()

  // 'status' will be 'success' immediately upon mount due to SSR hydration
  if (status === "pending") {
    return <div>Loading skeleton...</div> // Only seen if SSR fails or is skipped
  }

  if (status === "error") {
    return <div className="text-red-500">Failed to load feed.</div>
  }

  return (
    <div className="space-y-4">
      {data.pages.map((page, pageIndex) => (
        <div key={pageIndex} className="space-y-4">
          {page.edges.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </div>
      ))}

      {hasNextPage && (
        <button
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
          className="w-full rounded-lg py-3 text-sm font-semibold text-gray-600 transition-colors hover:bg-gray-100"
        >
          {isFetchingNextPage ? "Loading more..." : "Load more posts"}
        </button>
      )}
    </div>
  )
}
