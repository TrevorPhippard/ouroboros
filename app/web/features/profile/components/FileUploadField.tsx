"use client"

import { useEffect, useRef, useState } from "react"

export type UploadedImage = {
  key: string
  url: string
}

type FileUploadFieldProps = {
  initialValue?: string | null
  onUploadComplete?: (image: UploadedImage) => void
  accept?: string
  maxSizeMB?: number
}

type PresignedUpload = UploadedImage & {
  uploadUrl: string
  contentType: string
}

export default function FileUploadField({
  initialValue = "",
  onUploadComplete,
  accept = "image/jpeg,image/png,image/webp,image/gif",
  maxSizeMB = 10,
}: FileUploadFieldProps) {
  const [imageUrl, setImageUrl] = useState(initialValue || "")
  const [fileName, setFileName] = useState("")
  const [isDragging, setIsDragging] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState("")

  const inputRef = useRef<HTMLInputElement>(null)

  // Keep the preview synchronized if the parent loads or changes
  // the initial image asynchronously.
  useEffect(() => {
    setImageUrl(initialValue || "")
  }, [initialValue])

  function getFileName(url: string) {
    try {
      const pathname = new URL(url, window.location.origin).pathname
      return decodeURIComponent(pathname.split("/").pop() || "Image")
    } catch {
      return url.split("/").pop() || "Image"
    }
  }

  async function uploadFile(file?: File) {
    if (!file || isUploading) return

    setError("")

    if (!file.type.startsWith("image/")) {
      setError("Please select an image.")
      return
    }

    if (file.size > maxSizeMB * 1024 * 1024) {
      setError(`Image must be smaller than ${maxSizeMB} MB.`)
      return
    }

    setIsUploading(true)

    try {
      // 1. Ask our backend for a short-lived upload URL.
      const presignResponse = await fetch("/api/uploads/presign", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          contentType: file.type,
          size: file.size,
        }),
      })

      if (!presignResponse.ok) {
        throw new Error("Could not prepare the image upload.")
      }

      const upload: PresignedUpload = await presignResponse.json()

      // 2. Upload the file directly to S3.
      const uploadResponse = await fetch(upload.uploadUrl, {
        method: "PUT",
        headers: {
          "Content-Type": upload.contentType,
        },
        body: file,
      })

      if (!uploadResponse.ok) {
        throw new Error("Image upload failed. Please try again.")
      }

      // 3. Update the preview and notify the parent.
      setImageUrl(upload.url)
      setFileName(file.name)

      onUploadComplete?.({
        key: upload.key,
        url: upload.url,
      })
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Something went wrong uploading the image."
      )
    } finally {
      setIsUploading(false)

      if (inputRef.current) {
        inputRef.current.value = ""
      }
    }
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    setIsDragging(false)
    void uploadFile(e.dataTransfer.files[0])
  }

  if (imageUrl) {
    return (
      <div className="space-y-2">
        <img
          src={imageUrl}
          alt={fileName || getFileName(imageUrl)}
          className="h-40 w-40 rounded-lg border object-cover"
        />

        <p className="max-w-64 truncate text-sm text-gray-600">
          {fileName || getFileName(imageUrl)}
        </p>

        {error && (
          <p role="alert" className="text-sm text-red-600">
            {error}
          </p>
        )}

        <button
          type="button"
          disabled={isUploading}
          onClick={() => inputRef.current?.click()}
          className="text-sm font-medium text-blue-600 hover:underline disabled:opacity-50"
        >
          {isUploading ? "Uploading..." : "Replace image"}
        </button>

        <input
          ref={inputRef}
          type="file"
          accept={accept}
          className="hidden"
          onChange={(e) => void uploadFile(e.target.files?.[0])}
        />
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={(e) => void uploadFile(e.target.files?.[0])}
      />

      <div
        role="button"
        tabIndex={0}
        aria-label="Upload an image"
        aria-busy={isUploading}
        onClick={() => {
          if (!isUploading) inputRef.current?.click()
        }}
        onKeyDown={(e) => {
          if (!isUploading && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault()
            inputRef.current?.click()
          }
        }}
        onDragOver={(e) => {
          e.preventDefault()
          if (!isUploading) setIsDragging(true)
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
        className={[
          "flex min-h-44 flex-col items-center justify-center",
          "rounded-xl border-2 border-dashed p-6 text-center",
          "transition-colors",
          isDragging
            ? "border-blue-500 bg-blue-50"
            : "border-gray-300 hover:border-gray-400",
          isUploading ? "cursor-wait opacity-70" : "cursor-pointer",
        ].join(" ")}
      >
        {isUploading ? (
          <>
            <div className="mb-3 h-7 w-7 animate-spin rounded-full border-2 border-gray-300 border-t-blue-600" />
            <p className="font-medium">Uploading image...</p>
          </>
        ) : (
          <>
            <p className="font-medium text-gray-800">
              Drop an image here or click to browse
            </p>
            <p className="mt-1 text-sm text-gray-500">
              PNG, JPEG, WebP or GIF · Max {maxSizeMB} MB
            </p>
          </>
        )}
      </div>

      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
    </div>
  )
}
