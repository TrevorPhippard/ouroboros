import { z } from "zod"

export const connectionSchema = z.object({
  id: z.string().optional(),
})

export type Connection = z.infer<typeof connectionSchema>
