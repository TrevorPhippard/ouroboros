import { z } from "zod"

export const notificationSchema = z.object({
  id: z.string(),
  userId: z.string(),
  type: z.string(),
  actorId: z.string(),
  createdAt: z.string(),
  read: z.boolean(),
})

export const notificationsSchema = z.array(notificationSchema)

export type Notification = z.infer<typeof notificationSchema>
