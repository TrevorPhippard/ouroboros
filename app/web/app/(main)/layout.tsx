"use client"

import { Navbar } from "@/components/layout/Navbar"
import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useAuthStore } from "@/store/useAuthStore"

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  console.log("or here?")
  console.log("MainPage rendered")

  const router = useRouter()

  useEffect(() => {
    console.log("MainPage mounted")

    const token = useAuthStore.getState().token
    console.log("Token:", token)

    if (token) {
      console.log("Redirecting to feed")
      router.replace("/feed")
    } else {
      console.log("Redirecting to login")
      router.replace("/login")
    }
  }, [router])

  return (
    <>
      <Navbar />
      {children}
    </>
  )
}
