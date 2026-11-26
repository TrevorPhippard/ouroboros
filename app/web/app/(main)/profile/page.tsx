import { ProfileView } from "@/features/profile/components/ProfileView"
import { useAuthStore } from "@/store/useAuthStore"
import { Metadata } from "next"

export const metadata: Metadata = {
  title: "My Profile | AppName",
  description: "Manage your professional profile and experiences.",
}

export default function ProfilePage() {
  const userId = useAuthStore.getState().user?.id
  console.log("seregreger", userId)
  return (
    <main className="min-h-screen bg-gray-50 pt-16">
      <ProfileView userId={userId} />
    </main>
  )
}
