import { z } from "zod"

/**
 * The GraphQL Experience type. `id` is optional here because this schema is
 * also used by client-side forms before a new experience has been persisted.
 */
export const experienceSchema = z.object({
  id: z.string().optional(),
  title: z.string().min(1, "Title is required"),
  company: z.string().min(1, "Company is required"),
  startDate: z.string().min(1, "Start date is required"),
  endDate: z.string().nullable().optional(),
  description: z.string(),
})

/** Shape returned by the backend's `Profile` GraphQL type. */
export const profileSchema = z.object({
  id: z.string(),
  userId: z.string(),
  headline: z.string().nullable(),
  about: z.string().nullable(),
  displayName: z.string().nullable(),
  avatarUrl: z.string().nullable(),
  bio: z.string().nullable(),
  experiences: z.array(experienceSchema),
})

/** The subset of `User` selected by the profile-page query. */
export const userProfileDataSchema = z.object({
  id: z.string(),
  username: z.string(),
  displayName: z.string().nullable(),
  avatarUrl: z.string().nullable(),
  bio: z.string().nullable(),
  followersCount: z.number(),
  followingCount: z.number(),
  isFollowing: z.boolean(),
})

/** Matches the backend's `UpdateProfileInput` fields and nullability. */
export const profileUpdateInputSchema = z.object({
  headline: z.string().nullable().optional(),
  about: z.string().nullable().optional(),
  displayName: z.string().nullable().optional(),
  avatarUrl: z.string().nullable().optional(),
  bio: z.string().nullable().optional(),
})

/** UI form values are strings; empty optional fields are converted to null on submit. */
export const profileFormSchema = z.object({
  displayName: z.string().max(120).optional(),
  headline: z.string().max(120, "Headline must be 120 characters or fewer").optional(),
  about: z.string().max(2000, "About must be 2000 characters or fewer").optional(),
  bio: z.string().max(2000, "Bio must be 2000 characters or fewer").optional(),
})

export type ExperienceType = z.infer<typeof experienceSchema>
export type ProfileType = z.infer<typeof profileSchema>
export type UserProfileData = z.infer<typeof userProfileDataSchema>
export type ProfileUpdateInput = z.infer<typeof profileUpdateInputSchema>
export type ProfileFormValues = z.infer<typeof profileFormSchema>

/** View model combines User query data with Profile fields returned by updates. */
export type ProfileViewModel = {
  id: string
  userId: string
  username: string
  displayName: string | null
  avatarUrl: string | null
  bio: string | null
  headline: string | null
  about: string | null
  followersCount: number
  followingCount: number
  isFollowing: boolean
  experiences: ExperienceType[]
}
