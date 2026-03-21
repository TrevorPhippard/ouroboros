import { ProfileView } from "@/features/profile/components/ProfileView"
import { Metadata } from "next"

export const metadata: Metadata = {
  title: "My Profile | AppName",
  description: "Manage your professional profile and experiences.",
}

export default function ProfilePage() {
  // In a real app, you'd get the 'userId' from your auth session (e.g., NextAuth or your Zustand store's synced cookie)
  const currentUserId = "user-123"

  return (
    <main className="min-h-screen bg-gray-50 pt-16">
      <ProfileView userId={currentUserId} />
    </main>
  )
}
