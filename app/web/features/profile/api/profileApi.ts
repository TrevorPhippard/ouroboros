// src/features/profile/api/profileApi.ts
import { Profile } from "../schemas"

// 1. In-Memory Mock Database
// This persists state during your development session so updates actually reflect in the UI.
const mockProfileDb: Record<string, Profile> = {
  "user-123": {
    id: "user-123",
    headline: "Senior Web Developer",
    about:
      "Specializing in clean architecture, React ecosystems, and scalable event-driven systems.",
    experiences: [
      {
        id: "exp-1",
        title: "Frontend Architect",
        company: "Tech Innovations Inc.",
        startDate: "2023-01",
      },
    ],
  },
}

// Network latency simulator
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// 2. The Fetcher Contracts

export const getProfileApi = async (userId: string): Promise<Profile> => {
  await delay(800) // Simulate an 800ms API gateway response

  const profile = mockProfileDb[userId]
  if (!profile) throw new Error("Profile not found")

  return profile
}

export const updateProfileApi = async (
  userId: string,
  updates: Partial<Profile> = {}
): Promise<Profile> => {
  await delay(1000) // Simulate mutation latency
  console.log("Received profile update:", updates)
  // Optional: Uncomment the next line to test your optimistic UI rollbacks!
  // if (Math.random() > 0.5) throw new Error("Simulated 500 Internal Server Error");

  const existingProfile = mockProfileDb[userId]
  if (!existingProfile) throw new Error("Profile not found")

  // Merge updates
  const updatedProfile = { ...existingProfile, ...updates }
  mockProfileDb[userId] = updatedProfile

  return updatedProfile
}
