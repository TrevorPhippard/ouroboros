// src/features/profile/components/ProfileView.tsx
"use client"

import { useEffect, useState } from "react"
import { useProfile } from "../hooks/useProfile"
import { EditableHeadline } from "./EditableField"
import { useUpdateProfile } from "../hooks/useUpdateProfile"

export const ProfileView = ({ userId }: { userId: string }) => {
  const { data: profile, isLoading, isError } = useProfile(userId)
  const { mutate: updateProfile } = useUpdateProfile()
  const [about, setAbout] = useState("")

  useEffect(() => {
    if (profile) {
      setAbout(profile.about ?? "")
    }
  }, [profile])

  if (isLoading) return <ProfileSkeleton /> // Standard UX practice
  if (isError || !profile)
    return <div className="p-4 text-red-500">Failed to load profile.</div>

  return (
    <div className="mx-auto max-w-4xl space-y-6 py-8">
      {/* Top Card: Intro & Banner */}
      <div className="overflow-hidden rounded-lg bg-white shadow">
        <div className="relative h-32 bg-slate-300">
          <div className="absolute -bottom-12 left-6 h-24 w-24 overflow-hidden rounded-full border-4 border-white bg-white shadow-sm">
            <img
              src={
                profile.avatarUrl ||
                `https://api.dicebear.com/7.x/avataaars/svg?seed=${profile.id}`
              }
              alt="Avatar"
            />
          </div>
        </div>

        <div className="px-6 pt-16 pb-6">
          <h1 className="text-2xl font-bold text-gray-900">
            {profile.name || "Your Name Here"}
          </h1>
          <div className="mt-1 text-gray-600">
            <EditableHeadline initialValue={profile.headline} />
          </div>
          <p className="mt-2 text-sm text-gray-500">
            {profile.followingCount ?? 0} following • {profile.followersCount ?? 0} followers
          </p>
        </div>
      </div>

      {/* About Section */}
      <div className="rounded-lg bg-white p-6 shadow">
        <h2 className="mb-4 text-xl font-semibold text-gray-900">About</h2>
        <p className="whitespace-pre-wrap text-gray-700">
          {profile.about || "Add a summary about yourself..."}
        </p>
      </div>

      <div className="rounded-lg bg-white p-6 shadow">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-gray-900">Profile</h2>
          <button
            type="button"
            onClick={() => updateProfile({ about })}
            className="rounded-md bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700"
          >
            Save About
          </button>
        </div>
        <textarea
          value={about}
          onChange={(event) => setAbout(event.target.value)}
          className="min-h-32 w-full rounded-md border border-gray-200 p-3 text-sm text-gray-700"
        />
      </div>
    </div>
  )
}

// A quick loading skeleton to prevent layout shift
const ProfileSkeleton = () => (
  <div className="mx-auto max-w-4xl animate-pulse space-y-6 py-8">
    <div className="h-64 rounded-lg bg-gray-200 shadow"></div>
    <div className="h-40 rounded-lg bg-gray-200 shadow"></div>
    <div className="h-96 rounded-lg bg-gray-200 shadow"></div>
  </div>
)
