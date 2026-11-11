export default function Loading() {
  return (
    <main
      className="mx-auto max-w-6xl px-4 py-6"
      aria-label="Loading content"
      aria-busy="true"
    >
      <div className="grid grid-cols-1 gap-6 md:grid-cols-[220px_minmax(0,1fr)] lg:grid-cols-[220px_minmax(0,1fr)_280px]">
        {/* Left sidebar */}
        <aside className="hidden space-y-4 md:block">
          <div className="animate-pulse overflow-hidden rounded-lg border border-gray-200 bg-white">
            <div className="h-16 bg-gray-200" />
            <div className="space-y-3 p-4">
              <div className="mx-auto h-16 w-16 rounded-full bg-gray-200" />
              <div className="mx-auto h-4 w-3/4 rounded bg-gray-200" />
              <div className="mx-auto h-3 w-1/2 rounded bg-gray-200" />
              <div className="mt-5 border-t border-gray-100 pt-4">
                <div className="h-3 w-full rounded bg-gray-200" />
                <div className="mt-3 h-3 w-2/3 rounded bg-gray-200" />
              </div>
            </div>
          </div>
        </aside>

        {/* Main feed */}
        <section className="min-w-0 space-y-4">
          {/* Post composer skeleton */}
          <div className="animate-pulse rounded-lg border border-gray-200 bg-white p-4">
            <div className="flex items-center gap-3">
              <div className="h-12 w-12 shrink-0 rounded-full bg-gray-200" />
              <div className="h-11 flex-1 rounded-full bg-gray-100" />
            </div>
            <div className="mt-5 flex justify-between gap-3">
              <div className="h-4 w-16 rounded bg-gray-200" />
              <div className="h-4 w-16 rounded bg-gray-200" />
              <div className="h-4 w-16 rounded bg-gray-200" />
            </div>
          </div>

          {/* Feed post skeletons */}
          {[1, 2, 3].map((post) => (
            <article
              key={post}
              className="animate-pulse rounded-lg border border-gray-200 bg-white p-4"
            >
              <div className="flex items-center gap-3">
                <div className="h-12 w-12 shrink-0 rounded-full bg-gray-200" />
                <div className="flex-1 space-y-2">
                  <div className="h-3 w-1/3 rounded bg-gray-200" />
                  <div className="h-3 w-1/2 rounded bg-gray-100" />
                  <div className="h-3 w-1/4 rounded bg-gray-100" />
                </div>
              </div>

              <div className="mt-5 space-y-2">
                <div className="h-3 w-full rounded bg-gray-200" />
                <div className="h-3 w-5/6 rounded bg-gray-200" />
                <div className="h-3 w-2/3 rounded bg-gray-200" />
              </div>

              {post === 1 && (
                <div className="mt-4 h-48 rounded-md bg-gray-200" />
              )}

              <div className="mt-5 border-t border-gray-100 pt-4">
                <div className="h-4 w-full rounded bg-gray-100" />
              </div>
            </article>
          ))}
        </section>

        {/* Right sidebar */}
        <aside className="hidden space-y-4 lg:block">
          <div className="animate-pulse space-y-4 rounded-lg border border-gray-200 bg-white p-4">
            <div className="h-4 w-2/3 rounded bg-gray-200" />
            {[1, 2, 3].map((item) => (
              <div key={item} className="space-y-2">
                <div className="h-3 w-full rounded bg-gray-200" />
                <div className="h-3 w-4/5 rounded bg-gray-100" />
              </div>
            ))}
          </div>
        </aside>
      </div>
    </main>
  )
}
