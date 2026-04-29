import { SignInForm } from "@/features/auth/components/SignInForm"

import Link from "next/link"

export default function LoginPage() {
  return (
    <div className="flex min-h-screen flex-col items-center bg-white pt-12 md:bg-[#f3f2ef]">
      {/* Brand Logo */}
      <div className="mb-8 flex w-40 items-center gap-1">
        <img src="cleo.png" alt="" />
      </div>

      <div className="w-full max-w-[400px] bg-white p-8 md:rounded-lg md:shadow-md">
        <h1 className="mb-2 text-3xl font-semibold">Sign in</h1>
        <p className="mb-6 text-sm text-gray-600">
          Stay connected with your professional network.
        </p>
        <SignInForm />
      </div>

      <p className="mt-8 text-sm">
        New to ouroboros?{" "}
        <Link
          href="/register"
          className="font-semibold text-[#0a66c2] hover:underline"
        >
          Join now
        </Link>
      </p>
    </div>
  )
}
