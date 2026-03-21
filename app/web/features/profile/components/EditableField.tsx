import { useState } from "react"
import { useUpdateProfile } from "../hooks/useUpdateProfile"

export const EditableHeadline = ({
  initialValue,
}: {
  initialValue: string
}) => {
  const [isEditing, setIsEditing] = useState(false)
  const [value, setValue] = useState(initialValue)
  const { mutate, isPending } = useUpdateProfile()

  const handleSave = () => {
    mutate({ headline: value })
    setIsEditing(false)
  }

  if (isEditing) {
    return (
      <div className="flex items-center space-x-2">
        <input
          autoFocus
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="rounded border border-blue-500 px-2 py-1"
        />
        <button
          onClick={handleSave}
          className="rounded bg-blue-600 px-3 py-1 text-sm text-white"
        >
          Save
        </button>
        <button
          onClick={() => setIsEditing(false)}
          className="text-sm text-gray-500"
        >
          Cancel
        </button>
      </div>
    )
  }

  return (
    <div
      className={`group flex cursor-pointer items-center space-x-2 ${isPending ? "opacity-50" : ""}`}
      onClick={() => setIsEditing(true)}
    >
      <h2 className="text-xl font-semibold">{value}</h2>
      <span className="hidden text-sm text-blue-500 group-hover:inline">
        ✎ Edit
      </span>
    </div>
  )
}
