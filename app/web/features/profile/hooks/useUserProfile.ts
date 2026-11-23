// Keep the existing hook name for components that consume a public profile.
// Both profile hooks now share the same schema-valid GraphQL query and cache.
export { useProfile as useUserProfile } from "./useProfile"
