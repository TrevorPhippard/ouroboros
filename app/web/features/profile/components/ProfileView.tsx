// src/features/profile/components/ProfileView.tsx
"use client"

import { useProfile } from "../hooks/useProfile"
import { EditableHeadline } from "./EditableField"
import { ExperienceSection } from "./ExperienceSection"
import { Form } from "@/components/ui/form/Form"
import { profileSchema } from "../schemas"
import { useUpdateProfile } from "../hooks/useUpdateProfile"

export const ProfileView = ({ userId }: { userId: string }) => {
  const { data: profile, isLoading, isError } = useProfile(userId)
  const { mutate: updateProfile } = useUpdateProfile()

  if (isLoading) return <ProfileSkeleton /> // Standard UX practice
  if (isError || !profile)
    return <div className="p-4 text-red-500">Failed to load profile.</div>

  return (
    <div className="mx-auto max-w-4xl space-y-6 py-8">
      {/* Top Card: Intro & Banner */}
      <div className="overflow-hidden rounded-lg bg-white shadow">
        <div className="relative h-32 bg-slate-300">
          {/* Banner Placeholder */}
          <div className="absolute -bottom-12 left-6 h-24 w-24 overflow-hidden rounded-full border-4 border-white bg-white shadow-sm">
            <img
              src={`https://api.dicebear.com/7.x/avataaars/svg?seed=${profile.id}`}
              alt="Avatar"
            />
          </div>
        </div>

        <div className="px-6 pt-16 pb-6">
          <h1 className="text-2xl font-bold text-gray-900">{profile.name}</h1>{" "}
          {/* Mocked Name */}
          {/* Our Inline Editable Component */}
          <div className="mt-1 text-gray-600">
            {profile.headline || "Your headline goes here..."}
            <EditableHeadline initialValue={profile.headline} />
          </div>
          <p className="mt-2 text-sm text-gray-500">
            Toronto, Ontario, Canada • 500+ connections
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

      {/* Experience Section (Wrapped in our Form provider) */}
      <div className="rounded-lg bg-white p-6 shadow">
        {/* We wrap the complex array section in our Form component.
          When saved, it triggers the TanStack mutation optimistically.
        */}
        <Form
          schema={profileSchema}
          defaultValues={profile}
          onSubmit={(data) => updateProfile(data)}
        >
          <ExperienceSection />

          <div className="mt-6 flex justify-end">
            <button
              type="submit"
              className="rounded-md bg-blue-600 px-4 py-2 font-medium text-white transition-colors hover:bg-blue-700"
            >
              Save Experiences
            </button>
          </div>
        </Form>
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
