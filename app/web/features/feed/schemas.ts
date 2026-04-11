import { z } from "zod"

export const feedSchema = z.object({
  id: z.string().optional(),
})

export const postSchema = z.object({
  id: z.string(),
  author: z.object({
    id: z.string(),
    name: z.string().max(100),
    avatarUrl: z.string().url().optional(),
    headline: z.string().max(200).optional(),
  }),
  content: z.string().max(500),
  likes: z.number().min(0),
  hasLiked: z.boolean(),
  createdAt: z.string(),
})

export type FeedType = z.infer<typeof feedSchema>
export type PostType = z.infer<typeof postSchema>
