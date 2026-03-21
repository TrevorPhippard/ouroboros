"use client"

import { useQuery } from "@apollo/client"
import { GET_FEED } from "@/lib/queries"
import { MessageSquare, Heart, Share2 } from "lucide-react"

export default function PostFeed() {
  const { loading, error, data } = useQuery(GET_FEED, {
    pollInterval: 5000, // Refresh every 5 seconds for "real-time" feel
  })

  if (loading)
    return <div className="p-4 text-center">Loading your timeline...</div>
  if (error)
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
