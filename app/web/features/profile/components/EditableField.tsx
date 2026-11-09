"use client"

import { useState } from "react"
import { useUpdateProfile } from "../hooks/useUpdateProfile"

type EditableFieldProps = {
  label: string
  field: "name" | "headline" | "about"
  initialValue?: string | null
  multiline?: boolean
  placeholder?: string
}

export const EditableField = ({
  label,
  field,
  initialValue,
  multiline = false,
  placeholder,
}: EditableFieldProps) => {
  const [isEditing, setIsEditing] = useState(false)
  const [draft, setDraft] = useState("")
  const { mutateAsync, isPending } = useUpdateProfile()

  const displayValue = initialValue ?? ""

  const handleEdit = () => {
    setDraft(displayValue)
    setIsEditing(true)
  }

  const handleCancel = () => {
    setIsEditing(false)
    setDraft("")
  }

  const handleSave = async () => {
    if (draft === displayValue) {
      setIsEditing(false)
      return
    }

    try {
      await mutateAsync({ [field]: draft })
      setIsEditing(false)
      setDraft("")
    } catch {
      // Keep the editor open so the user can retry.
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
        <div
          className="cursor-pointer break-words whitespace-pre-wrap text-gray-800"
          onClick={handleEdit}
        >
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
