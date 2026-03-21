"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { executeGraphQL } from "@/services/graphql/client"

const postSchema = z.object({
  content: z
    .string()
    .min(1, "Post cannot be empty")
    .max(3000, "Post is too long"),
})

type PostFormValues = z.infer<typeof postSchema>

export const CreatePostForm = () => {
  const queryClient = useQueryClient()
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<PostFormValues>({
    resolver: zodResolver(postSchema),
  })

  const createPost = useMutation({
    mutationFn: (data: PostFormValues) =>
      executeGraphQL({
        query: `mutation CREATE_POST { ... }`,
        variables: data,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
      reset()
    },
  })

  return (
    <form
      onSubmit={handleSubmit((data) => createPost.mutate(data))}
      className="rounded-lg border bg-white p-4"
    >
      <textarea
        {...register("content")}
        placeholder="Start a post..."
        className="w-full resize-none p-2 text-sm outline-none"
        rows={3}
      />
      {errors.content && (
        <span className="text-xs text-red-500">{errors.content.message}</span>
      )}

      <div className="mt-2 flex justify-end">
        <button
          disabled={isSubmitting || createPost.isPending}
          className="rounded-full bg-blue-600 px-4 py-1.5 text-sm font-medium text-white disabled:opacity-50"
        >
          {createPost.isPending ? "Posting..." : "Post"}
        </button>
      </div>
    </form>
  )
}
