"use client"

import { useState } from "react"
import { useUpdateProfile } from "../hooks/useUpdateProfile"
import type { ProfileUpdateInput } from "../schemas"

type EditableFieldProps = {
  label: string
  field: keyof ProfileUpdateInput
  userId?: string
  initialValue?: string | null
  multiline?: boolean
  placeholder?: string
}

export const EditableField = ({
  label,
  field,
  userId,
  initialValue,
  multiline = false,
  placeholder,
}: EditableFieldProps) => {
  const [isEditing, setIsEditing] = useState(false)
  const [draft, setDraft] = useState("")
  const [saveError, setSaveError] = useState("")
  const { mutateAsync, isPending } = useUpdateProfile(userId)

  const displayValue = initialValue ?? ""

  const handleEdit = () => {
    setDraft(displayValue)
    setSaveError("")
    setIsEditing(true)
  }

  const handleCancel = () => {
    setIsEditing(false)
    setDraft("")
    setSaveError("")
  }

  const handleSave = async () => {
    if (draft === displayValue) {
      setIsEditing(false)
      return
    }

    try {
      const updates = { [field]: draft } as ProfileUpdateInput
      await mutateAsync(updates)
      setIsEditing(false)
      setDraft("")
      setSaveError("")
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : "Could not save this field.")
    }
  }

  return (
    <div className="group space-y-2">
      <div className="flex items-center justify-between gap-3">
        <label className="text-sm font-medium text-gray-500">{label}</label>

        {!isEditing && (
          <button
            type="button"
            onClick={handleEdit}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            Edit
          </button>
        )}
      </div>

      {isEditing ? (
        <div className="space-y-2">
          {multiline ? (
            <textarea
              autoFocus
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder={placeholder}
              disabled={isPending}
              className="min-h-32 w-full rounded-md border border-gray-300 p-3 text-sm focus:border-blue-500 focus:outline-none"
            />
          ) : (
            <input
              autoFocus
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              placeholder={placeholder}
              disabled={isPending}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            />
          )}

          {saveError && <p role="alert" className="text-sm text-red-600">{saveError}</p>}

          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleSave}
              disabled={isPending}
              className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isPending ? "Saving..." : "Save"}
            </button>

            <button
              type="button"
              onClick={handleCancel}
              disabled={isPending}
              className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <div className="break-words whitespace-pre-wrap text-gray-800">
          {displayValue || (
            <span className="text-gray-400">
              {placeholder ?? `Add ${label.toLowerCase()}...`}
            </span>
          )}
        </div>
      )}
    </div>
  )
}
