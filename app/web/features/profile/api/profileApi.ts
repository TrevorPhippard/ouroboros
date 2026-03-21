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

/**

// src/features/profile/api/profileApi.ts (FUTURE GRAPHQL STATE)
import { request, gql } from 'graphql-request';
import { Profile } from '../schema';

// This URL would likely come from your environment variables
const API_GATEWAY_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL || 'http://localhost:8080/query';

const GET_PROFILE_QUERY = gql`
  query GetProfile($userId: ID!) {
    profile(id: $userId) {
      id
      headline
      about
      experiences {
        id
        title
        company
        startDate
        endDate
      }
    }
  }
`;

const UPDATE_PROFILE_MUTATION = gql`
  mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
    updateProfile(userId: $userId, input: $input) {
      id
      headline
      about
    }
  }
`;

// Notice how the function signatures remain EXACTLY the same as the mock version
export const getProfileApi = async (userId: string): Promise<Profile> => {
  // If you are using an auth store (like Zustand), you would inject the Bearer token into headers here
  const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` };

  const data = await request<{ profile: Profile }>(
    API_GATEWAY_URL,
    GET_PROFILE_QUERY,
    { userId },
    headers
  );

  return data.profile;
};

export const updateProfileApi = async (userId: string, updates: Partial<Profile>): Promise<Profile> => {
  const headers = { Authorization: `Bearer ${localStorage.getItem('token')}` };

  const data = await request<{ updateProfile: Profile }>(
    API_GATEWAY_URL,
    UPDATE_PROFILE_MUTATION,
    { userId, input: updates },
    headers
  );

  return data.updateProfile;
};

 */
