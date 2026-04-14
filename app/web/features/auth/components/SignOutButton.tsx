"use client"

import { useSignOut } from "@/features/auth/hooks/useSignOut"
import { Button } from "@/components/ui/button"
import { LogOut } from "lucide-react"

export function SignOutButton() {
  const { mutate, isPending, isError, error } = useSignOut()

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        mutate()
      }}
    >
      <Button type="submit" variant="destructive" disabled={isPending}>
        <LogOut className="h-4 w-4" />
        {isPending ? "Signing out..." : "Sign Out"}
      </Button>

      {isError && (
        <p className="mt-2 text-sm text-destructive">{error.message}</p>
      )}
    </form>
  )
}
