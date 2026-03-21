"use client"

import { useState } from "react"
import { useMutation } from "@apollo/client"
import { CREATE_POST, GET_FEED } from "@/lib/queries"

export default function CreatePost() {
  const [content, setContent] = useState("")
  const [createPost, { loading }] = useMutation(CREATE_POST, {
    refetchQueries: [{ query: GET_FEED }], // Refresh the feed after posting
  })

  const handleSubmit = async () => {
    if (!content.trim()) return
    await createPost({ variables: { content } })
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
          disabled={loading || !content.trim()}
          className="rounded-full bg-blue-500 px-6 py-2 font-bold text-white transition hover:bg-blue-600 disabled:opacity-50"
        >
          {loading ? "Posting..." : "Post"}
        </button>
      </div>
    </div>
  )
}
