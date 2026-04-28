import { SignUpForm } from "@/features/auth/components/SignUpForm"

import Link from "next/link"

export default function RegisterPage() {
  return (
    <div className="flex min-h-screen flex-col items-center bg-white pt-12 md:bg-[#f3f2ef]">
      {/* Brand Logo */}
      <div className="mb-8 flex w-40 items-center gap-1">
        <img src="cleo.png" alt="" />
      </div>

      <div className="w-full max-w-[400px] bg-white p-8 md:rounded-lg md:shadow-md">
        <h1 className="mb-2 text-3xl font-semibold">Join now</h1>
        <p className="mb-6 text-sm text-gray-600">
          Build your professional profile.
        </p>
        <SignUpForm />
      </div>

      <p className="mt-8 text-sm">
        Already on ouroboros?{" "}
        <Link
          href="/login"
          className="font-semibold text-[#0a66c2] hover:underline"
        >
          Sign in
        </Link>
      </p>
    </div>
  )
}
