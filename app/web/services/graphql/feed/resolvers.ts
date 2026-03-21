import { Post } from "@/features/feed/hooks/useFeed"
import { subMinutes, subHours, subDays } from "date-fns"

// 1. Internal Mock State (Simulating a Database)
const MOCK_POSTS: Post[] = [
  {
    id: "p1",
    author: {
      id: "u1",
      name: "Watermelia",
      headline: "Flatulance Engineer @ Google",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=melon",
    },
    content: "Pfft!",
    likes: 1240,
    hasLiked: false,
    createdAt: subMinutes(new Date(), 45).toISOString(),
  },
  {
    id: "p2",
    author: {
      id: "u2",
      name: "Aplon",
      headline: "Worth of Love Researcher @ Apple",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=apple",
    },
    content: "You do not deserve love",
    likes: 850,
    hasLiked: true,
    createdAt: subHours(new Date(), 2).toISOString(),
  },
  {
    id: "p3",
    author: {
      id: "u3",
      name: "Pineaplon",
      headline: "Luxury Home Designer @ Facebook",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=pineapple",
    },
    content: "You are not worthy of this luxury home?",
    likes: 420,
    hasLiked: false,
    createdAt: subDays(new Date(), 1).toISOString(),
  },
]

// Helper to simulate "more data" for infinite scroll
const generateMorePosts = (cursor: string | null, limit: number) => {
  const startIndex = cursor
    ? MOCK_POSTS.findIndex((p) => p.id === cursor) + 1
    : 0
  const page = MOCK_POSTS.slice(startIndex, startIndex + limit)

  // In a real app, if we hit the end, we might generate procedurally or return empty
  const nextCursor = page.length > 0 ? page[page.length - 1].id : null

  return {
    edges: page,
    pageInfo: {
      nextCursor: startIndex + limit < MOCK_POSTS.length ? nextCursor : null,
      hasNextPage: startIndex + limit < MOCK_POSTS.length,
    },
  }
}

// 2. Resolver Implementation
export const feedResolvers = {
  // Query: Fetch feed with pagination
  getFeed: (variables: { cursor: string | null; limit?: number }) => {
    const limit = variables.limit || 2
    return generateMorePosts(variables.cursor, limit)
  },

  // Mutation: Like/Unlike logic
  likePost: (variables: { postId: string }) => {
    const post = MOCK_POSTS.find((p) => p.id === variables.postId)

    if (!post) {
      throw new Error("POST_NOT_FOUND")
    }

    // Toggle logic
    if (post.hasLiked) {
      post.likes -= 1
      post.hasLiked = false
    } else {
      post.likes += 1
      post.hasLiked = true
    }

    return post
  },

  // Mutation: Create Post (pushes to top of mock "DB")
  createPost: (variables: { content: string; author: any }) => {
    const newPost: Post = {
      id: `p-${Math.random().toString(36).substr(2, 9)}`,
      author: variables.author,
      content: variables.content,
      likes: 0,
      hasLiked: false,
      createdAt: new Date().toISOString(),
    }

    MOCK_POSTS.unshift(newPost) // Put at the top of the feed
    return newPost
  },
}
