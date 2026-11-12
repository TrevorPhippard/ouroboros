// import { gqlRequest } from "@/services/graphql/client"
// import { GET_FEED } from "@/lib/queries"
import { PostType } from "../schemas"
import { useAuthStore } from "@/store/useAuthStore"
// import { time } from "console"

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
      // const response = await gqlRequest({
      //   query: GET_FEED,
      //   variables: { userId, cursor: pageParam },
      // })

      const response = {
        feed: {
          nextCursor: pageParam === "cursor-2" ? null : "cursor-2",
          items: [
            {
              cursor: "cursor-1",
              post: {
                id: "post-1",
                content:
                  "Just shipped the new dashboard UI using Tailwind CSS and TanStack Query! Excited for user feedback. 🚀",
                createdAt: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 mins ago
                author: {
                  id: "profile-1",
                  username: "sarah_dev",
                  displayName: "Sarah Jenkins",
                  avatarUrl:
                    "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=150&h=150&fit=crop&crop=faces",
                  bio: "Frontend Engineer | UI/UX enthusiast",
                },
              },
            },
            {
              cursor: "cursor-2",
              post: {
                id: "post-2",
                content:
                  "Working on state management strategies today. What is everyone's current favorite setup?",
                createdAt: new Date(Date.now() - 1000 * 60 * 120).toISOString(), // 2 hours ago
                author: {
                  id: "profile-2",
                  username: "alex_codes",
                  displayName: "Alex Rivera",
                  avatarUrl:
                    "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150&h=150&fit=crop&crop=faces",
                  bio: "Full Stack Developer & Coffee Addict ☕",
                },
              },
            },
          ],
        },
      }

      return {
        feed: {
          items:
            response.feed?.items?.flatMap((item) => {
              if (!item.post?.author) return []
              return [
                {
                  id: item.post.id,
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
