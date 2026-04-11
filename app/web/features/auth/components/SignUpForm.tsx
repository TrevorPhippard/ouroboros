"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { signUpSchema, SignUpValues } from "@/features/auth/schemas"
import { useSignUp } from "@/features/auth/hooks/useSignUp"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Field, FieldLabel, FieldError } from "@/components/ui/field"

export function SignUpForm() {
  const { mutate, isPending, isError, error } = useSignUp()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignUpValues>({
    resolver: zodResolver(signUpSchema),
  })

  return (
    <form
      onSubmit={handleSubmit((data) => mutate(data))}
      className="w-full max-w-md space-y-6"
    >
      <Field>
        <FieldLabel htmlFor="name">Full Name</FieldLabel>
        <Input
          id="name"
          type="text"
          aria-invalid={!!errors.name}
          {...register("name")}
        />
        {errors.name && <FieldError>{errors.name.message}</FieldError>}
      </Field>

      <Field>
        <FieldLabel htmlFor="email">Email</FieldLabel>
        <Input
          id="email"
          type="email"
          aria-invalid={!!errors.email}
          {...register("email")}
        />
        {errors.email && <FieldError>{errors.email.message}</FieldError>}
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

      <Field>
        <FieldLabel htmlFor="confirmPassword">Confirm Password</FieldLabel>
        <Input
          id="confirmPassword"
          type="password"
          aria-invalid={!!errors.confirmPassword}
          {...register("confirmPassword")}
        />
        {errors.confirmPassword && (
          <FieldError>{errors.confirmPassword.message}</FieldError>
        )}
      </Field>

      <Button type="submit" disabled={isPending} className="w-full">
        {isPending ? "Creating account..." : "Sign Up"}
      </Button>

      {isError && (
        <p className="mt-2 text-sm text-destructive">{error.message}</p>
      )}
    </form>
  )
}
