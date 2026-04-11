"use client"

import React, { useEffect, useState } from "react"
import { useAuthStore } from "@/store/useAuthStore"
import { useRouter, usePathname } from "next/navigation"

export const RouteGuard = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useAuthStore((s: any) => s.isAuthenticated)
  const router = useRouter()
  const pathname = usePathname()
  const [isMounted, setIsMounted] = useState(false)

  useEffect(() => {
    setIsMounted(true)
  }, [])

  useEffect(() => {
    if (isMounted && !isAuthenticated && pathname !== "/login") {
      router.push("/login")
    }
  }, [isMounted, isAuthenticated, pathname, router])

  // Prevent flickering while checking auth state
  if (!isMounted) return null

  return <>{children}</>
}
