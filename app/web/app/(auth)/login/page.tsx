"use client"

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
        <p className="mb-6 text-sm text-gray-600">Dogs with jobs</p>
        <SignInForm />
        <div className="mt-6 flex items-center gap-2 text-gray-400">
          <div className="h-[1px] flex-1 bg-gray-300"></div>
          <span className="text-xs">or</span>
          <div className="h-[1px] flex-1 bg-gray-300"></div>
        </div>
        <button className="mt-6 w-full rounded-full border border-gray-600 py-2.5 font-semibold text-gray-600 transition-colors hover:bg-gray-50">
          Sign in with Poodle
        </button>
      </div>

      <p className="mt-8 text-sm">
        New to Puppo?{" "}
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
