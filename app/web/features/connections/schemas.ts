import { z } from "zod"

export const connectionSchema = z.object({
  id: z.string().optional(),
})

export const edgeUserSchema = z.object({
  id: z.string(),
  name: z.string(),
  username: z.string(),
  headline: z.string().optional(),
  avatarUrl: z.string().optional(),
})

export const recommendationSchema = z.array(edgeUserSchema)

export const userSchema = z.object({
  id: z.string(),
  name: z.string(),
  username: z.string(),
  headline: z.string().optional(),
  avatarUrl: z.string().optional(),
})

export type Connection = z.infer<typeof connectionSchema>
export type Recommendation = z.infer<typeof recommendationSchema>
export type User = z.infer<typeof userSchema>
export type EdgeUser = z.infer<typeof edgeUserSchema>
