"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Camera, X } from "lucide-react"
import { profileSchema, ProfileFormValues } from "../schemas"
import { useUpdateProfile } from "../hooks/useUpdateProfile"
import { useState } from "react"

export const EditProfileModal = ({
  initialData,
  onClose,
}: {
  initialData: any
  onClose: () => void
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const { mutateAsync: updateProfile, isPending } = useUpdateProfile()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: initialData,
  })

  const onSubmit = async (data: ProfileFormValues) => {
    await updateProfile({ data, avatarFile: selectedFile || undefined })
    onClose()
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-2xl overflow-hidden rounded-xl bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 className="text-xl font-semibold">Edit intro</h2>
          <button
            onClick={onClose}
            className="rounded-full p-1 hover:bg-gray-100"
          >
            <X />
          </button>
        </div>

        <form
          onSubmit={handleSubmit(onSubmit)}
          className="max-h-[80vh] space-y-4 overflow-y-auto p-6"
        >
          {/* Avatar Upload Mock */}
          <div
            className="group relative h-24 w-24 cursor-pointer"
            onClick={() => document.getElementById("avatar-input")?.click()}
          >
            <img
              src={
                selectedFile
                  ? URL.createObjectURL(selectedFile)
                  : initialData.avatarUrl
              }
              className="h-full w-full rounded-full border-4 border-white object-cover shadow"
              alt="Avatar"
            />
            <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/30 opacity-0 transition-opacity group-hover:opacity-100">
              <Camera className="h-8 w-8 text-white" />
            </div>
            <input
              id="avatar-input"
              type="file"
              hidden
              accept="image/*"
              onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm text-gray-600">
                First name*
              </label>
              <input
                {...register("firstName")}
                className="w-full rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black"
              />
              {errors.firstName && (
                <p className="mt-1 text-xs text-red-600">
                  {errors.firstName.message}
                </p>
              )}
            </div>
            <div>
              <label className="mb-1 block text-sm text-gray-600">
                Last name*
              </label>
              <input
                {...register("lastName")}
                className="w-full rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-600">
              Headline*
            </label>
            <textarea
              {...register("headline")}
              rows={2}
              className="w-full resize-none rounded border border-gray-400 px-3 py-1.5 outline-none focus:border-black"
            />
          </div>

          <div className="flex justify-end border-t pt-4">
            <button
              type="submit"
              disabled={isPending}
              className="rounded-full bg-[#0a66c2] px-6 py-1.5 font-semibold text-white transition-colors hover:bg-[#004182] disabled:opacity-50"
            >
              {isPending ? "Saving..." : "Save"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
