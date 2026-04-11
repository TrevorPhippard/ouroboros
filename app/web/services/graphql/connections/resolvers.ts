export const connectionResolvers = {
  getRecommendations: () => ({
    edges: [
      {
        id: "u1",
        name: "Watermelia",
        headline: "Flatulance Engineer @ Google",
        avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=melon",
      },
      {
        id: "u2",
        name: "Aplon",
        headline: "Worth of Love Researcher @ Apple",
        avatarUrl: "/avatars/satya.jpg",
      },
      {
        id: "u3",
        name: "Pineaplon",
        headline: "Luxury Home Designer @ Facebook",
        avatarUrl: "/avatars/sam.jpg",
      },
    ],
  }),
  sendRequest: (variables: { userId: string }) => {
    // Simulate a random 5% failure rate to test our rollback logic
    if (Math.random() < 0.05) throw new Error("CONNECTION_SERVICE_UNAVAILABLE")
    return { success: true, targetId: variables.userId }
  },
}
