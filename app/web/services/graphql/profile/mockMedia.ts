// Simulating a Media Microservice
export const mediaService = {
  getPresignedUrl: async (fileName: string) => {
    await new Promise((r) => setTimeout(r, 400))
    return {
      uploadUrl: `https://mock-s3-bucket.s3.amazonaws.com/${fileName}?signature=mock`,
      publicUrl: `https://cdn.linkedin-clone.com/uploads/${fileName}`,
    }
  },
  uploadToS3: async (file: File, url: string) => {
    // Simulate binary upload latency
    await new Promise((r) => setTimeout(r, 1200))
    return true
  },
}
