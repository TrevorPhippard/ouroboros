"use client"

import { useSendConnectionRequest } from "../hooks/useConnections"
import { UserPlus } from "lucide-react"

export const RecommendationCard = ({ user }: { user: any }) => {
  const { mutate: connect, isPending } = useSendConnectionRequest()

  return (
    <div className="flex flex-col items-center rounded-lg border border-gray-200 bg-white p-4 text-center shadow-sm">
      <img
        src={user.avatarUrl}
        className="mb-2 h-16 w-16 rounded-full"
        alt=""
      />
      <h3 className="cursor-pointer text-sm font-semibold hover:underline">
        {user.name}
      </h3>
      <p className="mb-4 line-clamp-2 h-8 text-xs text-gray-500">
        {user.headline}
      </p>

      <button
        onClick={() => connect(user.id)}
        disabled={isPending}
        className="flex w-full items-center justify-center gap-1 rounded-full border border-blue-600 px-4 py-1 text-sm font-semibold text-blue-600 transition-colors hover:bg-blue-50 disabled:opacity-50"
      >
        <UserPlus className="h-4 w-4" />
        {isPending ? "Connecting..." : "Connect"}
      </button>
    </div>
  )
}
