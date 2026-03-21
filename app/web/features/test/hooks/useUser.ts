import { useQuery } from "@tanstack/react-query"
import { graphqlClient } from "@/lib/graphqlClient"
import { gql } from "graphql-request"

const GET_USER = gql`
  query {
    users(ids: ["1", "2"]) {
      id
      mockData1
    }
    todos(ids: ["A", "B"]) {
      id
      mockData2
    }
  }
`

export function useUser(id: string) {
  return useQuery({
    queryKey: ["user", id],
    queryFn: async () => {
      console.log("Fetching user data for id:", id) // Debug log
      const data = await graphqlClient.request(GET_USER, { id })
      return data
    },
    staleTime: 1000 * 60, // 1 min cache
  })
}
