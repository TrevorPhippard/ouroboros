"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/useAuthStore";

export default function MainPage() {
  const router = useRouter();

  useEffect(() => {
    // .getState() reads the value once without triggering React renders
    const token = useAuthStore.getState().token;
    console.log("Token on main page:", token);
    if (token) {
      router.push("/feed");
    } else {
      router.push("/login");
    }
  }, [router]);

  // Next.js will cleanly render this loading state on the server,
  // and the effect will immediately route on the client. No mismatches!
  return (
    <div className="flex h-screen items-center justify-center">
      <p>Loading...</p>
    </div>
  );
}