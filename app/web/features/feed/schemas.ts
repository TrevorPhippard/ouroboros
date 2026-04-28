import { z } from "zod"

export const feedAuthorSchema = z.object({
  id: z.string(),
  username: z.string(),
  name: z.string(),
  avatarUrl: z.string().url().optional(),
  headline: z.string().optional(),
})

export const postSchema = z.object({
  id: z.string(),
  author: feedAuthorSchema,
  content: z.string().max(500),
  likes: z.number().min(0),
  hasLiked: z.boolean(),
  createdAt: z.string(),
})

export const feedSchema = z.object({
  items: z.array(postSchema),
  nextCursor: z.string().nullable().optional(),
})

export type FeedType = z.infer<typeof feedSchema>
export type PostType = z.infer<typeof postSchema>
