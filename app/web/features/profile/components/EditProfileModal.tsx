"use client"

import { useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { X } from "lucide-react"
import { profileFormSchema, type ProfileFormValues, type ProfileUpdateInput } from "../schemas"
import { useUpdateProfile } from "../hooks/useUpdateProfile"
import FileUploadField, { type UploadedImage } from "./FileUploadField"

type EditableProfileData = {
  displayName: string | null
  headline: string | null
  about: string | null
  bio: string | null
  avatarUrl: string | null
}

export const EditProfileModal = ({
  initialData,
  onClose,
  userId,
}: {
  initialData: EditableProfileData
  onClose: () => void
  userId?: string
}) => {
  const [uploadedAvatar, setUploadedAvatar] = useState<UploadedImage | null>(null)
  const [saveError, setSaveError] = useState("")
  const { mutateAsync: updateProfile, isPending } = useUpdateProfile(userId)

  const {
    register,
    handleSubmit,
    formState: { errors, dirtyFields },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: {
      displayName: initialData.displayName ?? "",
      headline: initialData.headline ?? "",
      about: initialData.about ?? "",
      bio: initialData.bio ?? "",
    },
  })

  const onSubmit = async (data: ProfileFormValues) => {
    setSaveError("")
    const updates: ProfileUpdateInput = {}
    // Only send fields the user changed. The supplied schema cannot read
    // Profile-only fields on page load, so submitting untouched empty defaults
    // would accidentally clear values already stored by the backend.
    if (dirtyFields.displayName) updates.displayName = data.displayName?.trim() || null
    if (dirtyFields.headline) updates.headline = data.headline?.trim() || null
    if (dirtyFields.about) updates.about = data.about?.trim() || null
    if (dirtyFields.bio) updates.bio = data.bio?.trim() || null
    if (uploadedAvatar) updates.avatarUrl = uploadedAvatar.url

    if (Object.keys(updates).length === 0) {
      onClose()
      return
    }

    try {
      await updateProfile(updates)
      onClose()
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : "Could not save your profile.")
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-2xl overflow-hidden rounded-xl bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 className="text-xl font-semibold">Edit profile</h2>
          <button type="button" onClick={onClose} aria-label="Close edit profile" className="rounded-full p-1 hover:bg-gray-100">
            <X />
          </button>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="max-h-[80vh] space-y-4 overflow-y-auto p-6">
          <div>
            <label className="mb-1 block text-sm text-gray-600">Profile avatar</label>
            <FileUploadField initialValue={initialData.avatarUrl} onUploadComplete={setUploadedAvatar} />
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-600">Display name</label>
            <input {...register("displayName")} className="w-full rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black" />
            {errors.displayName && <p className="mt-1 text-xs text-red-600">{errors.displayName.message}</p>}
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-600">Headline</label>
            <input {...register("headline")} className="w-full rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black" />
            {errors.headline && <p className="mt-1 text-xs text-red-600">{errors.headline.message}</p>}
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-600">Bio</label>
            <textarea {...register("bio")} rows={3} className="w-full resize-y rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black" />
            {errors.bio && <p className="mt-1 text-xs text-red-600">{errors.bio.message}</p>}
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-600">About</label>
            <textarea {...register("about")} rows={4} className="w-full resize-y rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black" />
            {errors.about && <p className="mt-1 text-xs text-red-600">{errors.about.message}</p>}
          </div>

          {saveError && <p role="alert" className="text-sm text-red-600">{saveError}</p>}

          <div className="flex justify-end border-t pt-4">
            <button type="submit" disabled={isPending} className="rounded-full bg-[#0a66c2] px-6 py-1.5 font-semibold text-white transition-colors hover:bg-[#004182] disabled:opacity-50">
              {isPending ? "Saving..." : "Save"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
