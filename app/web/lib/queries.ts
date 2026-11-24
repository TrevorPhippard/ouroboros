/**
 * GraphQL operations are kept beside the profile feature so their selected
 * fields and variables can be checked against the backend schema together.
 */
// export const GET_PROFILE = /* GraphQL */ `
//   query GetProfile($userId: ID!) {
//     profile(userId: $userId) {
//       id
//       userId
//       headline
//       about
//       displayName
//       avatarUrl
//       bio
//       experiences {
//         id
//         title
//         company
//         startDate
//         endDate
//         description
//       }
//     }
//   }
// `

export const GET_PROFILE = gql`
  query Profile($userId: ID!) {
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

// export const UPDATE_PROFILE = /* GraphQL */ `
//   mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
//     updateProfile(userId: $userId, input: $input) {
//       id
//       userId
//       headline
//       about
//       displayName
//       avatarUrl
//       bio
//       experiences {
//         id
//         title
//         company
//         startDate
//         endDate
//         description
//       }
//     }
//   }
// `

export const UPDATE_PROFILE = gql`
  mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
    updateProfile(userId: $userId, input: $input) {
      id
      headline
      about
      username
      displayName
      avatarUrl
      bio
      followersCount
      followingCount
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
import { gql } from "graphql-request"

export const GET_FEED = gql`
  query GetFeed($userId: ID!, $cursor: String) {
    feed(userId: $userId, limit: 10, cursor: $cursor) {
      items {
        postId
        cursor
        post {
          id
          content
          createdAt
          author {
            id
            username
            displayName
            avatarUrl
            bio
          }
        }
      }
    }
  }
`

export const CREATE_POST = gql`
  mutation CreatePost($input: CreatePostInput!) {
    createPost(input: $input) {
      id
      content
      createdAt
      UserId
    }
  }
`

export const LIKE_POST = gql`
  mutation LikePost($postId: ID!) {
    likePost(postId: $postId) {
      id
      content
      createdAt
      UserId
    }
  }
`

export const GET_NOTIFICATIONS = gql`
  query GetNotifications($userId: ID!) {
    notifications(userId: $userId, limit: 20) {
      id
      userId
      type
      actorId
      createdAt
      read
    }
  }
`

export const GET_RECOMMENDATIONS = gql`
  query GetRecommendations {
    recommendations {
      id
      username
      displayName
      avatarUrl
      bio
    }
  }
`

export const SEND_CONNECT = gql`
  mutation SendConnect($userId: ID!) {
    sendConnect(userId: $userId) {
      success
    }
  }
`

export const SIGN_UP = gql`
  mutation SignUp($input: SignUpInput!) {
    signUp(input: $input) {
      token
      user {
        id
        email
        username
        displayName
      }
    }
  }
`

export const SIGN_IN = gql`
  mutation SignIn($input: SignInInput!) {
    signIn(input: $input) {
      token
      user {
        id
        email
        username
        displayName
      }
    }
  }
`

export const SIGN_OUT = gql`
  mutation SignOut {
    signOut {
      success
    }
  }
`
