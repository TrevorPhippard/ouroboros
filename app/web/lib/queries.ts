import { gql } from "graphql-request"

export const GET_FEED = gql`
  query GetFeed($userId: ID!) {
    feed(userId: $userId, limit: 10) {
      items {
        postId
        cursor
        post {
          content
          author {
            username
            avatarUrl
          }
        }
      }
    }
  }
`
export const GET_PROFILE = gql`
  query GetProfile($userId: ID!) {
    profile(id: $userId) {
      id
      headline
      about
      experiences {
        id
        title
        company
        startDate
        endDate
      }
    }
  }
`

export const UPDATE_PROFILE = gql`
  mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
    updateProfile(userId: $userId, input: $input) {
      id
      headline
      about
    }
  }
`

export const CREATE_POST = gql`
  mutation CreatePost($content: String!) {
    createPost(content: $content) {
      id
      content
    }
  }
`

export const LIKE_POST = gql`
  mutation LikePost($postId: ID!) {
    likePost(postId: $postId) {
      id
      likes
    }
  }
`

export const GET_UNREAD_NOTIFICATIONS = gql`
  query GetUnreadNotifications($userId: ID!) {
    unreadNotifications(userId: $userId) {
      id
      content
      createdAt
    }
  }
`

export const GET_RECOMMENDATIONS = gql`
  query GetRecommendations {
    recommendations {
      id
      displayName
      headline
      avatarUrl
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
      id
      displayName
      email
    }
  }
`

export const SIGN_IN = gql`
  mutation SignIn($input: SignInInput!) {
    signIn(input: $input) {
      token
      user {
        id
        displayName
        email
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
