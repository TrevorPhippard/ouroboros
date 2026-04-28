"use client"

import { useInfiniteQuery } from "@tanstack/react-query"
import { useFeedQuery } from "../hooks/useFeed"
import { PostCard } from "./PostCard"

export const FeedList = () => {
  const queryOptions = useFeedQuery()
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, status } =
    useInfiniteQuery(queryOptions)

  // 'status' will be 'success' immediately upon mount due to SSR hydration
  if (status === "pending") {
    return (
      <div>
        Loading skeleton... {data}
        {status}
      </div>
    ) // Only seen if SSR fails or is skipped
  }

  if (status === "error") {
    return <div className="text-red-500">Failed to load feed.</div>
  }
  // Debug log to inspect the structure of 'data'
  return (
    <div className="space-y-4">
      {data.pages.map((page, index) => (
        <div key={index} className="space-y-4">
          {page.feed.items.map((post) => (
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
