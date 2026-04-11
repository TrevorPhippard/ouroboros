## GetFeed

```
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
```

## Variables

```
{
  "userId": "user_123"
}
```

## GetProfile

```
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
```

## UpdateProfile

```
mutation UpdateProfile($userId: ID!, $input: UpdateProfileInput!) {
  updateProfile(userId: $userId, input: $input) {
    id
    headline
    about
  }
}
```

## CreatePost

```
mutation CreatePost($content: String!) {
  createPost(content: $content) {
    id
    content
  }
}
```

## LikePost

```
mutation LikePost($postId: ID!) {
  likePost(postId: $postId) {
    id
    likes
  }
}
```

## GetUnreadNotifications

```
query GetUnreadNotifications($userId: ID!) {
  unreadNotifications(userId: $userId) {
    id
    content
    createdAt
  }
}
```

## GetRecommendations

```
query GetRecommendations {
  recommendations {
    id
    displayName
    headline
    avatarUrl
  }
}
```

## SendConnect

```
mutation SendConnect($userId: ID!) {
  sendConnect(userId: $userId) {
    success
  }
}
```

## SignUp

```
mutation SignUp($input: SignUpInput!) {
  signUp(input: $input) {
    id
    displayName
    email
  }
}
```

## SignIn

```
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

```
mutation SignOut {
  signOut {
    success
  }
}
```

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

```

mutation SignOut {
signOut {
success
}
}

```
 signOut {
    success
  }
}

```
