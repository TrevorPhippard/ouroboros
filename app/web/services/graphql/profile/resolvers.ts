const MOCK_PROFILES: Record<string, any> = {
  u1: {
    id: "u1",
    name: "Watermelia",
    headline: "Flatulance Engineer @ Google",
    avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=melon",
    coverUrl: "https://images.unsplash.com/photo-1579546929518-9e396f3cc809",
    about:
      "Empowering every person and every organization on the planet to achieve more.",
    followers: 10400000,
    connections: "500+",
    location: "Redmond, Washington, United States",
  },
  // Add other mock profiles as needed...
}

export const profileResolvers = {
  getUserProfile: (variables: { id: string }) => {
    const profile = MOCK_PROFILES[variables.id]
    if (!profile) throw new Error("PROFILE_NOT_FOUND")
    return profile
  },
}
