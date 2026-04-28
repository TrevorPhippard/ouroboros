## GetFeed

```graphql
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
```

## Variables

```json
{
  "userId": "user-1"
}
```

## GetProfile

```graphql
query GetProfile($userId: ID!) {
  user(id: $userId) {
    id
    username
    displayName
    avatarUrl
    bio
    followersCount
    followingCount
  }
}
```

## UpdateProfile

```graphql
mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
  updateProfile(userId: $userId, input: $input) {
    id
    headline
    about
  }
}
```

## CreatePost

```graphql
mutation CreatePost($input: CreatePostInput!) {
  createPost(input: $input) {
    id
    content
    createdAt
    authorId
  }
}
```

## LikePost

```graphql
mutation LikePost($postId: ID!) {
  likePost(postId: $postId) {
    id
    content
    createdAt
    authorId
  }
}
```

## GetNotifications

```graphql
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
```

## GetRecommendations

```graphql
query GetRecommendations {
  recommendations {
    id
    username
    displayName
    avatarUrl
    bio
  }
}
```

## SendConnect

```graphql
mutation SendConnect($userId: ID!) {
  sendConnect(userId: $userId) {
    success
  }
}
```

## SignUp

```graphql
mutation SignUp($input: SignUpInput!) {
  signUp(input: $input) {
    id
    displayName
    email
  }
}
```

## SignIn

```graphql
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
```

## SignOut

```graphql
mutation SignOut {
  signOut {
    success
  }
}
```
