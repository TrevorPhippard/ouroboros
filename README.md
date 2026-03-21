# ouroboros

query CheckMyFeed {
feed(userId: "user_01", limit: 5) {
items {
postId
cursor
post {
id
content
createdAt
author {
username
displayName
}
}
}
}
}
