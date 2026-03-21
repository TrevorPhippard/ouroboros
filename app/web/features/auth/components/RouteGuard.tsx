"use client"

import { useAuthStore } from "@/store/useAuthStore"
import { useRouter, usePathname } from "next/navigation"
import { useEffect, useState } from "react"

export const RouteGuard = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuthStore()
  const router = useRouter()
  const pathname = usePathname()
  const [isMounted, setIsMounted] = useState(false)

  // useEffect(() => {
  //   setIsMounted(true)
  // }, [])

  useEffect(() => {
    if (isMounted && !isAuthenticated && pathname !== "/login") {
      router.push("/login")
    }
  }, [isMounted, isAuthenticated, pathname, router])

  // Prevent flickering while checking auth state
  if (!isMounted) return null

  return <>{children}</>
}
