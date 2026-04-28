import { z } from "zod"

export const experienceSchema = z.object({
  id: z.string().optional(),
  title: z.string().min(2, "Title is required"),
  company: z.string().min(2, "Company is required"),
  startDate: z.string(),
  endDate: z.string().optional(),
})

export const profileSchema = z.object({
  id: z.string(),
  name: z.string(),
  avatarUrl: z.string().optional(),
  headline: z.string().min(5).max(120),
  about: z.string().max(2000).optional(),
  followersCount: z.number().optional(),
  followingCount: z.number().optional(),
  experiences: z.array(experienceSchema).default([]),
})

export type ExperienceType = z.infer<typeof experienceSchema>
export type ProfileType = z.infer<typeof profileSchema>
