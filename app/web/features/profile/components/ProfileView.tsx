"use client"

import { useState } from "react"
import { useProfile } from "../hooks/useProfile"
import { useUpdateProfile } from "../hooks/useUpdateProfile"
import { EditableField } from "./EditableField"
import FileUploadField, { type UploadedImage } from "./FileUploadField"
type Props = {
  profile: {
    avatarUrl: string | null
    avatarKey: string | null
  }
}

export const ProfileView = ({ userId }: { userId: string }) => {
  const { data: profile, isLoading, isError } = useProfile(userId)
  const { mutate: updateProfile } = useUpdateProfile()
  const [avatar, setAvatar] = useState<UploadedImage | null>()

  if (isLoading) return <ProfileSkeleton />

  if (isError || !profile) {
    return <div className="p-4 text-red-500">Failed to load profile.</div>
  }

  console.log(profile)
  return (
    <div className="mx-auto max-w-4xl space-y-6 py-8">
      {/* Intro and banner */}
      <div className="overflow-hidden rounded-lg bg-white shadow">
        <div className="relative h-32 bg-slate-300">
          <div className="absolute -bottom-12 left-6 h-24 w-24 overflow-hidden rounded-full border-4 border-white bg-white shadow-sm">
            <img
              src={
                profile.avatarUrl ||
                `https://api.dicebear.com/7.x/avataaars/svg?seed=${profile.id}`
              }
              alt="Avatar"
              className="h-full w-full object-cover"
            />
          </div>
        </div>

        <div className="space-y-4 px-6 pt-16 pb-6">
          <EditableField
            label="Name"
            field="name"
            initialValue={profile.name}
            placeholder="Your name"
          />

          <EditableField
            label="Headline"
            field="headline"
            initialValue={profile.headline}
            placeholder="Add a professional headline"
          />

          <p className="text-sm text-gray-500">
            {profile.followingCount ?? 0} following
            {" • "}
            {profile.followersCount ?? 0} followers
          </p>
        </div>
      </div>

      {/* About */}
      <div className="rounded-lg bg-white p-6 shadow">
        <h2 className="mb-4 text-xl font-semibold text-gray-900">About Me</h2>

        <EditableField
          label="Description"
          field="about"
          initialValue={profile.about}
          multiline
          placeholder="Add a summary about yourself..."
        />
      </div>

      {/* Additional profile settings */}
      <div className="space-y-4 rounded-lg bg-white p-6 shadow">
        <h2 className="text-xl font-semibold text-gray-900">
          Profile settings
        </h2>

        <form
          onSubmit={async (e) => {
            e.preventDefault()
            await updateProfile({ about })
          }}
        >
          <label className="mb-2 block font-medium">Profile avatar</label>

          <FileUploadField
            initialValue={profile.avatarUrl}
            onUploadComplete={setAvatar}
          />

          <button
            type="button"
            className="rounded-md bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700"
          >
            Save About
          </button>
        </form>
      </div>
    </div>
  )
}

const ProfileSkeleton = () => (
  <div className="mx-auto max-w-4xl animate-pulse space-y-6 py-8">
    <div className="h-64 rounded-lg bg-gray-200 shadow" />
    <div className="h-40 rounded-lg bg-gray-200 shadow" />
    <div className="h-96 rounded-lg bg-gray-200 shadow" />
  </div>
)
