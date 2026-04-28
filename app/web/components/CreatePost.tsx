"use client"

import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { gqlRequest } from "@/services/graphql/client"
import { CREATE_POST } from "@/lib/queries"
import { useAuthStore } from "@/store/useAuthStore"

export default function CreatePost() {
  const [content, setContent] = useState("")
  const userId = useAuthStore((state) => state.user?.id ?? "user-1")
  const queryClient = useQueryClient()

  const { mutateAsync, isPending } = useMutation({
    mutationFn: (content: string) =>
      gqlRequest({
        query: CREATE_POST,
        variables: {
          input: {
            authorId: userId,
            content,
          },
        },
      }),

    onSuccess: () => {
      // 🔁 replaces refetchQueries
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })

  const handleSubmit = async () => {
    if (!content.trim()) return

    await mutateAsync(content)
    setContent("")
  }

  return (
    <div className="mb-6 rounded-xl border border-gray-200 bg-white p-4 shadow-sm">
      <textarea
        className="w-full resize-none border-none text-xl outline-none focus:ring-0"
        placeholder="What's happening?"
        rows={3}
        value={content}
        onChange={(e) => setContent(e.target.value)}
      />
      <div className="flex justify-end border-t pt-3">
        <button
          onClick={handleSubmit}
          disabled={isPending || !content.trim()}
          className="rounded-full bg-blue-500 px-6 py-2 font-bold text-white transition hover:bg-blue-600 disabled:opacity-50"
        >
          {isPending ? "Posting..." : "Post"}
        </button>
      </div>
    </div>
  )
}
