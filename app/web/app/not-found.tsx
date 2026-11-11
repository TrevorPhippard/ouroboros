import Link from "next/link"

export default function NotFound() {
  return (
    <main className="flex min-h-[70vh] items-center justify-center px-4 py-16">
      <div className="w-full max-w-md text-center">
        <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-blue-50">
          <svg
            aria-hidden="true"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            className="h-10 w-10 text-blue-600"
          >
            <circle cx="11" cy="11" r="7" />
            <path
              strokeLinecap="round"
              d="m16 16 4 4M8.5 10h.01M13.5 10h.01M8.5 14c1.5-1.5 3.5-1.5 5 0"
            />
          </svg>
        </div>

        <p className="text-sm font-semibold tracking-widest text-blue-600">
          404
        </p>

        <h1 className="mt-3 text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Page not found
        </h1>

        <p className="mt-3 text-sm leading-6 text-gray-600">
          This page may have been moved, or the link might be incorrect.
          Let&apos;s get you back to your network.
        </p>

        <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
          <Link
            href="/"
            className="inline-flex items-center justify-center rounded-full bg-blue-600 px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-blue-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            Go to feed
          </Link>

          <button
            type="button"
            onClick={() => window.history.back()}
            className="inline-flex items-center justify-center rounded-full border border-gray-300 bg-white px-6 py-3 text-sm font-semibold text-gray-700 transition-colors hover:bg-gray-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            Go back
          </button>
        </div>
      </div>
    </main>
  )
}
