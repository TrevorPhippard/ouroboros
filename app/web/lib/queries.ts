import { gql } from "@apollo/client"

export const GET_FEED = gql`
  query GetFeed {
    feed {
      id
      content
      createdAt
      author {
        id
        displayName
        avatarUrl
      }
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
