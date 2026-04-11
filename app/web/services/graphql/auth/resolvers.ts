interface SignInInput {
  email: string
  password: string
}

interface SignUpInput {
  email: string
  password: string
  displayName: string
}

export const authResolvers = {
  signIn: async (_: unknown, { input }: { input: SignInInput }) => {
    const res = await fetch("http://localhost:4000/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    })

    return res.json()
  },
  signUp: async (_: unknown, { input }: { input: SignUpInput }) => {
    const res = await fetch("http://localhost:4000/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    })

    return res.json()
  },
  signOut: async () => {
    // In a real implementation, you might want to invalidate the token on the server
    return { success: true }
  },
}
