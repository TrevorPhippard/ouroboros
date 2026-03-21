import { z } from "zod"

export const feedSchema = z.object({
  id: z.string().optional(),
})

export type Feed = z.infer<typeof feedSchema>
