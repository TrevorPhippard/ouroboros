import { useMutation, useQueryClient } from "@tanstack/react-query"
import { graphqlClient } from "@/lib/graphqlClient"
import { gql } from "graphql-request"

const CREATE_POST = gql`
  mutation CreatePost($input: CreatePostInput!) {
    createPost(input: $input) {
      id
      content
      createdAt
    }
  }
`

export function useCreatePost() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: { content: string }) => {
      const data = await graphqlClient.request(CREATE_POST, { input })
      return data.createPost
    },

    // 🔥 KEY PART: invalidate queries like Apollo cache updates
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["feed"] })
    },
  })
}
