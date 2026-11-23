/**
 * GraphQL operations are kept beside the profile feature so their selected
 * fields and variables can be checked against the backend schema together.
 */
export const GET_PROFILE = /* GraphQL */ `
  query GetProfile($userId: ID!) {
    profile(userId: $userId) {
      id
      userId
      headline
      about
      displayName
      avatarUrl
      bio
      experiences {
        id
        title
        company
        startDate
        endDate
        description
      }
    }
  }
`

export const UPDATE_PROFILE = /* GraphQL */ `
  mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
    updateProfile(userId: $userId, input: $input) {
      id
      userId
      headline
      about
      displayName
      avatarUrl
      bio
      experiences {
        id
        title
        company
        startDate
        endDate
        description
      }
    }
  }
`

export const FOLLOW_USER = /* GraphQL */ `
  mutation FollowUser($userId: ID!) {
    followUser(userId: $userId) {
      success
    }
  }
`

export const UNFOLLOW_USER = /* GraphQL */ `
  mutation UnfollowUser($userId: ID!) {
    unfollowUser(userId: $userId) {
      success
    }
  }
`
