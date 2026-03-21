import { z } from "zod"

export const notificationSchema = z.object({
  id: z.string().optional(),
})

export type Notification = z.infer<typeof notificationSchema>
