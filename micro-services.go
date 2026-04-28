Notes


diagrams

context: user interact with app, ux

	pages w/ services required:

		feed:
			actions: 				load more, like, comment
			excluded: 			any sort of filtering
			services req: 	feed, posts, profile

		profile page:
			actions: 				update, upload image, view connections, connection statuses, view own posts
			excluded: 			logging of changes, undo
			services req: 	posts, profile, connections

			alt profile page
				actions: 			view, make friend request
				excluded: 		block user, view users connections
				services req: posts, profile

		Navbar:
			action: 				nav to feed, profile, notifications, logout
			services req:		auth

		Auth page:
			actions: 				sign in, sign up
			excluded: 			forget password
			services req:		auth


container: application layers, with you thinking

Kafka Events: need to write full scheme

Redis Keys:

components: parts of services, probably best for feed

code: code level implimentation

order of services to build

read/write databases

neo4j

hexagonal archutecture vs golang standard folder structure

why golang rational

why graphql
why gRPC
encryption

DAU amount, through put, target latancy, rate limiting

roles: authentication, authorization

ux flow

wireframes

api

caching where

security

deployment
observability




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



mutation CreatePost($input: CreatePostInput!) {
  createPost(input: $input) {
    id
    content
    createdAt
  }
}


{
  "input": {
    "authorId": "USER_ID_HERE",
    "content": "Post created with variables"
  }
}