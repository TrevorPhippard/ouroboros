"use client"

import React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useLogin } from "@/features/auth/hooks/useLogin"
import Link from "next/link"

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z.string().min(6, "Password must be at least 6 characters"),
})

type LoginFormValues = z.infer<typeof loginSchema>

export default function LoginPage() {
  const { mutate: login, isPending, error } = useLogin()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginFormValues) => {
    login(data)
  }

  return (
    <div className="flex min-h-screen flex-col items-center bg-white pt-12 md:bg-[#f3f2ef]">
      {/* Brand Logo */}
      <div className="mb-8 flex items-center gap-1">
        <span className="text-3xl font-bold tracking-tighter text-[#0a66c2]">
          PuppedIn
        </span>
        <div className="rounded-sm bg-[#0a66c2] px-1 text-2xl font-bold text-white">
          Pup
        </div>
      </div>

      <div className="w-full max-w-[400px] bg-white p-8 md:rounded-lg md:shadow-md">
        <h1 className="mb-2 text-3xl font-semibold">Sign in</h1>
        <p className="mb-6 text-sm text-gray-600">Dogs with jobs</p>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {/* Email Field */}
          <div>
            <input
              {...register("email")}
              placeholder="Email"
              className={`w-full border ${
                errors.email ? "border-red-600" : "border-gray-500"
              } rounded-md px-3 py-3 transition-all outline-none focus:border-[#0a66c2] focus:ring-1 focus:ring-[#0a66c2]`}
            />
            {errors.email && (
              <p className="mt-1 text-xs text-red-600">
                {errors.email.message}
              </p>
            )}
          </div>

          {/* Password Field */}
          <div>
            <input
              {...register("password")}
              type="password"
              placeholder="Password"
              className={`w-full border ${
                errors.password ? "border-red-600" : "border-gray-500"
              } rounded-md px-3 py-3 transition-all outline-none focus:border-[#0a66c2] focus:ring-1 focus:ring-[#0a66c2]`}
            />
            {errors.password && (
              <p className="mt-1 text-xs text-red-600">
                {errors.password.message}
              </p>
            )}
          </div>

          {error && (
            <div className="rounded-md bg-red-50 p-3 text-sm text-red-600">
              {(error as Error).message}
            </div>
          )}

          <button
            type="submit"
            disabled={isPending}
            className="w-full rounded-full bg-[#0a66c2] py-3 font-semibold text-white transition-colors hover:bg-[#004182] disabled:opacity-50"
          >
            {isPending ? "Signing in..." : "Sign in"}
          </button>
        </form>

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
        New to PuppedIn?{" "}
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
