## 1. Authentication

### Login Mutation

#### GraphQL

```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    token
    user {
      id
      email
    }
  }
}
```

#### JSON

```json
{
  "input": {
    "email": "user@example.com",
    "password": "securePassword123"
  }
}
```

---

### Register Mutation

#### GraphQL

```graphql
mutation Register($input: RegisterInput!) {
  register(input: $input) {
    token
    user {
      id
      email
    }
  }
}
```

#### JSON

```json
{
  "input": {
    "email": "newuser@example.com",
    "password": "securePassword123",
    "name": "Jane Doe"
  }
}
```

---

### Refresh Token Mutation

#### GraphQL

```graphql
mutation RefreshToken($token: String!) {
  refreshToken(token: $token) {
    token
    user {
      id
    }
  }
}
```

#### JSON

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn..."
}
```

---

### Logout Mutation

#### GraphQL

```graphql
mutation Logout {
  logout
}
```

---

## 2. User & Profile

### Get Current Authenticated User (`me`) Query

#### GraphQL

```graphql
query GetMe {
  me {
    id
    email
    profile {
      bio
      avatarUrl
    }
  }
}
```

---

### Get Specific User Query

#### GraphQL

```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    email
    profile {
      bio
      avatarUrl
    }
  }
}
```

#### JSON

```json
{
  "id": "user_12345"
}
```

---

### Get Profile Query

#### GraphQL

```graphql
query GetProfile($userId: ID!) {
  profile(userId: $userId) {
    bio
    avatarUrl
    metadata
  }
}
```

#### JSON

```json
{
  "userId": "user_12345"
}
```

---

### Update User Mutation

#### GraphQL

```graphql
mutation UpdateUser($input: UpdateUserInput!) {
  updateUser(input: $input) {
    id
    email
  }
}
```

#### JSON

```json
{
  "input": {
    "email": "updated@example.com"
  }
}
```

---

### Update Profile Mutation

#### GraphQL

```graphql
mutation UpdateProfile($input: UpdateProfileInput!) {
  updateProfile(input: $input) {
    bio
    avatarUrl
  }
}
```

#### JSON

```json
{
  "input": {
    "bio": "Software engineer and tech enthusiast.",
    "avatarUrl": "[https://example.com/avatar.jpg](https://example.com/avatar.jpg)"
  }
}
```

---

## 3. Posts

### Get Single Post Query

#### GraphQL

```graphql
query GetPost($id: ID!) {
  post(id: $id) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "id": "post_98765"
}
```

---

### List Posts Query

#### GraphQL

```graphql
query GetPosts($limit: Int, $offset: Int) {
  posts(limit: $limit, offset: $offset) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "limit": 10,
  "offset": 0
}
```

---

### Get User Posts Query

#### GraphQL

```graphql
query GetUserPosts($userId: ID!) {
  userPosts(userId: $userId) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "userId": "user_12345"
}
```

---

### Create Post Mutation

#### GraphQL

```graphql
mutation CreatePost($input: CreatePostInput!) {
  createPost(input: $input) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "input": {
    "title": "My First Post",
    "content": "Hello world! This is a test post content."
  }
}
```

---

### Update Post Mutation

#### GraphQL

```graphql
mutation UpdatePost($id: ID!, $input: UpdatePostInput!) {
  updatePost(id: $id, input: $input) {
    id
    title
    content
  }
}
```

#### JSON

```json
{
  "id": "post_98765",
  "input": {
    "title": "Updated Title",
    "content": "Updated content text."
  }
}
```

---

### Delete Post Mutation

#### GraphQL

```graphql
mutation DeletePost($id: ID!) {
  deletePost(id: $id)
}
```

#### JSON

```json
{
  "id": "post_98765"
}
```

---

## 4. Feed

### Home Feed Query

#### GraphQL

```graphql
query GetFeed($limit: Int, $offset: Int) {
  feed(limit: $limit, offset: $offset) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "limit": 10,
  "offset": 0
}
```

---

### Explore Feed Query

#### GraphQL

```graphql
query GetExploreFeed($limit: Int, $offset: Int) {
  exploreFeed(limit: $limit, offset: $offset) {
    id
    title
    content
    createdAt
  }
}
```

#### JSON

```json
{
  "limit": 10,
  "offset": 0
}
```

---

## 5. Connections

### Get Followers Query

#### GraphQL

```graphql
query GetFollowers($userId: ID!) {
  followers(userId: $userId) {
    id
    email
  }
}
```

#### JSON

```json
{
  "userId": "user_12345"
}
```

---

### Get Following Query

#### GraphQL

```graphql
query GetFollowing($userId: ID!) {
  following(userId: $userId) {
    id
    email
  }
}
```

#### JSON

```json
{
  "userId": "user_12345"
}
```

---

### Follow User Mutation

#### GraphQL

```graphql
mutation FollowUser($userId: ID!) {
  followUser(userId: $userId)
}
```

#### JSON

```json
{
  "userId": "user_67890"
}
```

---

### Unfollow User Mutation

#### GraphQL

```graphql
mutation UnfollowUser($userId: ID!) {
  unfollowUser(userId: $userId)
}
```

#### JSON

```json
{
  "userId": "user_67890"
}
```

---

## 6. Notifications

### Get Notifications Query

#### GraphQL

```graphql
query GetNotifications {
  notifications {
    id
    content
    read
    createdAt
  }
}
```

---

### Get Unread Notification Count Query

#### GraphQL

```graphql
query GetUnreadNotificationCount {
  unreadNotificationCount
}
```

---

### Mark Notification As Read Mutation

#### GraphQL

```graphql
mutation MarkNotificationAsRead($id: ID!) {
  markNotificationAsRead(id: $id) {
    id
    read
  }
}
```

#### JSON

```json
{
  "id": "notif_11122"
}
```

---

### Mark All Notifications As Read Mutation

#### GraphQL

```graphql
mutation MarkAllNotificationsAsRead {
  markAllNotificationsAsRead
}
```


### Get Unread Notification Count Query

#### GraphQL

```graphql
query GetUnreadNotificationCount {
  unreadNotificationCount
}
```

---

### Mark Notification As Read Mutation

#### GraphQL

```graphql
mutation MarkNotificationAsRead($id: ID!) {
  markNotificationAsRead(id: $id) {
    id
    read
  }
}
```

#### JSON

```json
{
  "id": "notif_11122"
}
```

---

### Mark All Notifications As Read Mutation

#### GraphQL

```graphql
mutation MarkAllNotificationsAsRead {
  markAllNotificationsAsRead
}
```
