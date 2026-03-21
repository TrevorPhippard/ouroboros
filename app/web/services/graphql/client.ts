import { useAuthStore } from "@/store/useAuthStore"
import { feedResolvers } from "./feed/resolvers"
import { profileResolvers } from "./profile/resolvers"

// Simulating a GraphQL operation signature
export interface GraphQLOperation<TVariables = any> {
  query: string
  variables?: TVariables
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// Central mocked gateway. Swappable to real fetch/GraphQL client later.
export const executeGraphQL = async <TData>(
  operation: GraphQLOperation,
  options: { simulateError?: boolean; delayMs?: number } = {}
): Promise<TData> => {
  const { delayMs = 600, simulateError = false } = options

  await sleep(delayMs)

  const token = useAuthStore.getState().token
  if (!token && operation.query.includes("protected_")) {
    throw new Error("UNAUTHENTICATED: Invalid JWT token")
  }

  if (simulateError) {
    throw new Error("INTERNAL_SERVER_ERROR: Downstream gRPC service timeout")
  }

  // Basic query router (Mocking a GraphQL Server)
  if (operation.query.includes("GET_FEED")) {
    return feedResolvers.getFeed(operation.variables) as TData
  }
  if (operation.query.includes("LIKE_POST")) {
    return feedResolvers.likePost(operation.variables) as TData
  }

  if (operation.query.includes("GET_USER_PROFILE")) {
    return profileResolvers.getUserProfile(operation.variables) as TData
  }

  throw new Error(`Unrecognized mocked operation: ${operation.query}`)
}
