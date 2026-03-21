import { QueryClient, defaultShouldDehydrateQuery } from "@tanstack/react-query"

// Safely creates a new QueryClient per request on the server
export function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // Standard SSR defaults
        staleTime: 60 * 1000, // 1 minute
      },
      dehydrate: {
        // Include pending queries in dehydration
        shouldDehydrateQuery: (query) =>
          defaultShouldDehydrateQuery(query) ||
          query.state.status === "pending",
      },
    },
  })
}

let browserQueryClient: QueryClient | undefined = undefined

export function getQueryClient() {
  if (typeof window === "undefined") {
    // Server: always make a new query client
    return makeQueryClient()
  } else {
    // Client: make a new query client if we don't already have one
    if (!browserQueryClient) browserQueryClient = makeQueryClient()
    return browserQueryClient
  }
}
