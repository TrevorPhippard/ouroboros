"use client"

import { useQuery } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { MessageSquare, Heart, Share2 } from "lucide-react"
import { CREATE_POST } from "@/lib/queries"
import { PostType } from "@/features/feed/schemas"
// import { PostType } from

export default function PostFeed() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["feed"],
    queryFn: () =>
      gqlRequest({
        query: CREATE_POST,
        variables: {
          content: "Hello World!",
        },
      }),
    refetchInterval: 5000, // replaces pollInterval
  })

  if (isLoading)
    return <div className="p-4 text-center">Loading your timeline...</div>

  if (error instanceof Error)
    return (
      <div className="p-4 text-red-500">
        Error loading feed: {error.message}
      </div>
    )

  return (
    <div className="space-y-4">
      {data.feed.map((post: any) => (
        <div
          key={post.id}
          className="rounded-xl border border-gray-200 bg-white p-4 shadow-sm transition hover:bg-gray-50"
        >
          <div className="mb-2 flex items-center space-x-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-500 font-bold text-white">
              {post.author.displayName[0]}
            </div>
            <div>
              <p className="font-bold text-gray-900">
                {post.author.displayName}
              </p>
              <p className="text-xs text-gray-500">
                {new Date(parseInt(post.createdAt)).toLocaleString()}
              </p>
            </div>
          </div>
          <p className="mb-4 text-lg text-gray-800">{post.content}</p>
          <div className="flex justify-between border-t pt-3 text-gray-500">
            <button className="flex items-center space-x-2 hover:text-blue-500">
              <MessageSquare size={18} /> <span>0</span>
            </button>
            <button className="flex items-center space-x-2 hover:text-red-500">
              <Heart size={18} /> <span>0</span>
            </button>
            <button className="flex items-center space-x-2 hover:text-green-500">
              <Share2 size={18} /> <span>0</span>
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
