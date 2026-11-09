// Optional in-memory mock used for isolated UI development. Production hooks
// use the schema-valid GraphQL operations in `profileQueries.ts`.
import type { ProfileType, ProfileUpdateInput } from "../schemas"

const mockProfileDb: Record<string, ProfileType> = {
  "user-123": {
    id: "profile-123",
    userId: "user-123",
    displayName: "Morgan Lee",
    avatarUrl: null,
    headline: "Senior Web Developer",
    about: "Specializing in clean architecture and scalable web applications.",
    bio: "Specializing in clean architecture and scalable web applications.",
    experiences: [
      {
        id: "exp-1",
        title: "Frontend Architect",
        company: "Tech Innovations Inc.",
        startDate: "2023-01",
        endDate: null,
        description: "Led frontend architecture and component-system development.",
      },
    ],
  },
}

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

export const getProfileApi = async (userId: string): Promise<ProfileType> => {
  await delay(800)

  const profile = mockProfileDb[userId]
  if (!profile) throw new Error("Profile not found")

  return profile
}

export const updateProfileApi = async (
  userId: string,
  updates: ProfileUpdateInput = {},
): Promise<ProfileType> => {
  await delay(1000)

  const existingProfile = mockProfileDb[userId]
  if (!existingProfile) throw new Error("Profile not found")

  const updatedProfile: ProfileType = { ...existingProfile, ...updates }
  mockProfileDb[userId] = updatedProfile
  return updatedProfile
}
