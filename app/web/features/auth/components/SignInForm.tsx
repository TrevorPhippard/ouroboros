"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { signInSchema, type SignInValues } from "@/features/auth/schemas"
import { useSignIn } from "@/features/auth/hooks/useSignIn"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "@/components/ui/field"

export function SignInForm() {
  const { mutate, isPending, isError, error } = useSignIn()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignInValues>({
    resolver: zodResolver(signInSchema),
  })

  return (
    <form
      onSubmit={handleSubmit((data) => mutate(data))}
      className="w-full max-w-md space-y-6"
    >
      <Field>
        <FieldLabel htmlFor="email">Email</FieldLabel>
        <Input
          id="email"
          type="email"
          aria-invalid={!!errors.email}
          {...register("email")}
        />
        {errors.email ? (
          <FieldError>{errors.email.message}</FieldError>
        ) : (
          <FieldDescription>Enter your account email.</FieldDescription>
        )}
      </Field>

      <Field>
        <FieldLabel htmlFor="password">Password</FieldLabel>
        <Input
          id="password"
          type="password"
          aria-invalid={!!errors.password}
          {...register("password")}
        />
        {errors.password && <FieldError>{errors.password.message}</FieldError>}
      </Field>

      <Button type="submit" disabled={isPending} className="w-full">
        {isPending ? "Signing in..." : "Sign In"}
      </Button>

      {isError && (
        <p className="mt-2 text-sm text-destructive">{error.message}</p>
      )}
    </form>
  )
}
