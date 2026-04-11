export async function gqlRequest({
  query,
  variables,
}: {
  query: string
  variables?: Record<string, unknown>
}) {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null

  const res = await fetch("http://localhost:4000/query", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: token ? `Bearer ${token}` : "",
    },
    body: JSON.stringify({ query, variables }),
  })

  const json = await res.json()

  if (json.errors) {
    throw new Error(json.errors[0].message)
  }

  return json.data
}
