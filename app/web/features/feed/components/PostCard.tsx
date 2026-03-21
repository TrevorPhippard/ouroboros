"use client"

import React, { useState } from "react"
import {
  ThumbsUp,
  MessageSquare,
  Repeat2,
  Send,
  MoreHorizontal,
  Globe,
} from "lucide-react"
import { Post, useLikePost } from "../hooks/useFeed"
import { cn } from "@/lib/utils"
import { formatDistanceToNow } from "date-fns"
import Link from "next/link"

export const PostCard = ({ post }: { post: Post }) => {
  const { mutate: likePost } = useLikePost()
  const [isExpanded, setIsExpanded] = useState(false)

  // Content truncation logic (LinkedIn standard is ~200 characters)
  const shouldTruncate = post.content.length > 200 && !isExpanded
  const displayedContent = shouldTruncate
    ? `${post.content.substring(0, 200)}...`
    : post.content

  return (
    <article className="mb-2 rounded-lg border border-gray-200 bg-white shadow-sm">
      {/* Header: Author Info */}
      <div className="flex items-start justify-between p-3 pb-2">
        <div className="flex gap-2">
          <Link href={`/profile/${post.author.id}`}>
            <img
              src={post.author.avatarUrl}
              alt={post.author.name}
              className="h-12 w-12 cursor-pointer rounded-full object-cover transition-opacity hover:opacity-80"
            />
          </Link>
          <div className="flex flex-col">
            <Link
              href={`/profile/${post.author.id}`}
              className="w-fit cursor-pointer text-sm font-semibold hover:text-blue-600 hover:underline"
            >
              {post.author.name}
            </Link>
            <p className="line-clamp-1 text-xs text-gray-500">
              {post.author.headline}
            </p>
            <div className="flex items-center gap-1 text-xs text-gray-400">
              <span>{formatDistanceToNow(new Date(post.createdAt))}</span>
              <span>•</span>
              <Globe className="h-3 w-3" />
            </div>
          </div>
        </div>
        <button className="rounded-full p-1 text-gray-600 transition-colors hover:bg-gray-100">
          <MoreHorizontal className="h-5 w-5" />
        </button>
      </div>

      {/* Content Body */}
      <div className="px-3 pb-2 text-sm break-words text-gray-800">
        <p>
          {displayedContent}
          {shouldTruncate && (
            <button
              onClick={() => setIsExpanded(true)}
              className="ml-1 font-semibold text-gray-500 hover:text-blue-600"
            >
              ...see more
            </button>
          )}
        </p>
      </div>

      {/* Social Metrics */}
      {post.likes > 0 && (
        <div className="flex items-center justify-between border-b border-gray-100 px-3 py-2">
          <div className="flex items-center gap-1">
            <div className="rounded-full bg-blue-100 p-0.5">
              <ThumbsUp className="h-2.5 w-2.5 fill-blue-600 text-blue-600" />
            </div>
            <span className="cursor-pointer text-xs text-gray-500 hover:text-blue-600 hover:underline">
              {post.likes}
            </span>
          </div>
          {/* Mocked comment count */}
          <span className="cursor-pointer text-xs text-gray-500 hover:text-blue-600 hover:underline">
            12 comments
          </span>
        </div>
      )}

      {/* Action Buttons */}
      <div className="flex items-center gap-1 px-1 py-1">
        <ActionButton
          icon={ThumbsUp}
          label="Like"
          active={post.hasLiked}
          onClick={() => likePost(post.id)}
        />
        <ActionButton icon={MessageSquare} label="Comment" />
        <ActionButton icon={Repeat2} label="Repost" />
        <ActionButton icon={Send} label="Send" />
      </div>
    </article>
  )
}

// Internal Sub-component for DRY action buttons
const ActionButton = ({
  icon: Icon,
  label,
  active,
  onClick,
}: {
  icon: any
  label: string
  active?: boolean
  onClick?: () => void
}) => (
  <button
    onClick={onClick}
    className={cn(
      "flex flex-1 items-center justify-center gap-2 rounded-md py-3 transition-colors hover:bg-gray-100",
      active ? "font-semibold text-blue-600" : "text-gray-500"
    )}
  >
    <Icon className={cn("h-5 w-5", active && "fill-blue-600")} />
    <span className="text-sm font-medium">{label}</span>
  </button>
)
