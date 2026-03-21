"use client"

import { useUserProfile } from "../hooks/useUserProfile"
import { UserPlus, Send } from "lucide-react"

export const ProfileView = ({ userId }: { userId: string }) => {
  const { data: profile, isLoading, isError } = useUserProfile(userId)

  if (isLoading)
    return (
      <div className="p-8 text-center text-gray-500">Loading profile...</div>
    )
  if (isError || !profile)
    return (
      <div className="p-8 text-center text-red-500">Profile not found.</div>
    )

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
      {/* Cover Photo */}
      <div className="relative h-48 w-full bg-gray-200">
        <img
          src={profile.coverUrl}
          alt="Cover"
          className="h-full w-full object-cover"
        />
      </div>

      <div className="relative px-6 pb-6">
        {/* Avatar positioned halfway over the cover photo */}
        <div className="absolute -top-16 rounded-full border-4 border-white bg-white">
          <img
            src={profile.avatarUrl}
            alt={profile.name}
            className="h-32 w-32 rounded-full object-cover"
          />
        </div>

        {/* Profile Info */}
        <div className="pt-20">
          <h1 className="text-2xl font-bold">{profile.name}</h1>
          <p className="mt-1 text-lg text-gray-800">{profile.headline}</p>
          <p className="mt-2 text-sm text-gray-500">
            {profile.location} •{" "}
            <span className="cursor-pointer font-semibold text-blue-600 hover:underline">
              {profile.connections} connections
            </span>
          </p>
          <p className="mt-1 text-sm text-gray-500">
            {new Intl.NumberFormat("en-US").format(profile.followers)} followers
          </p>

          {/* Action Buttons */}
          <div className="mt-4 flex gap-2">
            <button className="flex items-center gap-2 rounded-full bg-[#0a66c2] px-5 py-1.5 font-semibold text-white transition-colors hover:bg-[#004182]">
              <UserPlus className="h-4 w-4" /> Connect
            </button>
            <button className="flex items-center gap-2 rounded-full border border-[#0a66c2] px-5 py-1.5 font-semibold text-[#0a66c2] transition-colors hover:bg-blue-50">
              <Send className="h-4 w-4" /> Message
            </button>
          </div>
        </div>

        {/* About Section */}
        {profile.about && (
          <div className="mt-8 rounded-lg bg-gray-50 p-4">
            <h2 className="mb-2 text-xl font-semibold">About</h2>
            <p className="text-sm whitespace-pre-wrap text-gray-700">
              {profile.about}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
