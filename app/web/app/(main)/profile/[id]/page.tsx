import { dehydrate, HydrationBoundary } from "@tanstack/react-query"
import { getQueryClient } from "@/lib/getQueryClient"
import { executeGraphQL } from "@/services/graphql/client"
import { ProfileView } from "@/features/profile/components/ProfileView"
import { UserProfile } from "@/features/profile/hooks/useUserProfile"

// 1. Update the type definition so TypeScript knows params is a Promise
type Props = {
  params: Promise<{ id: string }>
}

export default async function UserProfilePage({ params }: Props) {
  // 2. Await the params object (Next.js 15 standard)
  const resolvedParams = await params
  const userId = resolvedParams.id

  const queryClient = getQueryClient()

  // 3. Use the unwrapped userId for your query prefetching
  await queryClient.prefetchQuery({
    queryKey: ["profile", userId],
    queryFn: () =>
      executeGraphQL<UserProfile>({
        query: `query GET_USER_PROFILE($id: ID!) { ... }`,
        variables: { id: userId },
      }),
  })

  return (
    <main className="mx-auto max-w-4xl py-6">
      <HydrationBoundary state={dehydrate(queryClient)}>
        {/* Pass the unwrapped userId to the Client Component */}
        <ProfileView userId={userId} />
      </HydrationBoundary>
    </main>
  )
}
