import { GraphQLClient } from "graphql-request"

const endpoint = "http://localhost:4000/query"

export const graphqlClient = new GraphQLClient(endpoint, {
  credentials: `include`,
  mode: `cors`,
})
