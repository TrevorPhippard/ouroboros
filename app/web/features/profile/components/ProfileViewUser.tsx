"use client"

import { UserPlus } from "lucide-react"
import { useUserProfile } from "../hooks/useUserProfile"
import { useToggleFollow } from "../hooks/useToggleFollow"

export const ProfileView = ({ userId }: { userId: string }) => {
  const { data: profile, isLoading, isError } = useUserProfile(userId)
  const {
    mutate: toggleFollow,
    isPending: isUpdatingFollow,
    error: followError,
  } = useToggleFollow(userId)

  if (isLoading) {
    return <div className="p-8 text-center text-gray-500">Loading profile...</div>
  }
  if (isError || !profile) {
    return <div className="p-8 text-center text-red-500">Profile not found.</div>
  }

  const displayName = profile.displayName ?? profile.username
  const avatarUrl = profile.avatarUrl || `https://api.dicebear.com/7.x/avataaars/svg?seed=${profile.id}`

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
      <div className="relative h-48 w-full bg-gray-200">
        <img src={avatarUrl} alt={`${displayName}'s profile cover`} className="h-full w-full object-cover" />
      </div>

      <div className="relative px-6 pb-6">
        <div className="absolute -top-16 rounded-full border-4 border-white bg-white">
          <img src={avatarUrl} alt={displayName} className="h-32 w-32 rounded-full object-cover" />
        </div>

        <div className="pt-20">
          <h1 className="text-2xl font-bold">{displayName}</h1>
          {profile.headline && <p className="mt-1 text-lg text-gray-800">{profile.headline}</p>}
          <p className="mt-2 text-sm text-gray-500">
            <span className="font-semibold">{profile.followingCount}</span> following{" • "}
            <span className="font-semibold">{profile.followersCount}</span> followers
          </p>

          <div className="mt-4 flex gap-2">
            <button
              type="button"
              disabled={isUpdatingFollow}
              onClick={() => toggleFollow(!profile.isFollowing)}
              className="flex items-center gap-2 rounded-full bg-[#0a66c2] px-5 py-1.5 font-semibold text-white transition-colors hover:bg-[#004182] disabled:opacity-50"
            >
              <UserPlus className="h-4 w-4" />
              {isUpdatingFollow ? "Saving..." : profile.isFollowing ? "Unfollow" : "Follow"}
            </button>
          </div>

          {followError && <p role="alert" className="mt-2 text-sm text-red-600">{followError.message}</p>}
        </div>

        {profile.bio && (
          <div className="mt-8 rounded-lg bg-gray-50 p-4">
            <h2 className="mb-2 text-xl font-semibold">About</h2>
            <p className="text-sm whitespace-pre-wrap text-gray-700">{profile.bio}</p>
          </div>
        )}
      </div>
    </div>
  )
}
