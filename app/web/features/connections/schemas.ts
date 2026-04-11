import { z } from "zod"

export const connectionSchema = z.object({
  id: z.string().optional(),
})

export const edgeUserSchema = z.object({
  id: z.string(),
  name: z.string(),
  headline: z.string(),
  mutualConnections: z.number(),
})

export const recommendationSchema = z.object({
  id: z.string().optional(),
  edges: z.array(edgeUserSchema),
})

export const userSchema = z.object({
  id: z.string().optional(),
  name: z.string(),
  headline: z.string(),
  avatarUrl: z.string(),
  coverUrl: z.string(),
  about: z.string(),
  followers: z.number(),
  connections: z.string(),
  location: z.string(),
})

export type Connection = z.infer<typeof connectionSchema>
export type Recommendation = z.infer<typeof recommendationSchema>
export type User = z.infer<typeof userSchema>
export type EdgeUser = z.infer<typeof edgeUserSchema>
