"use client"

import React, { useState } from "react"
import Link from "next/link"
import { useAuthStore } from "@/store/useAuthStore"
import {
  Home,
  Users,
  Briefcase,
  MessageSquare,
  Bell,
  Search,
  ChevronDown,
  LogOut,
} from "lucide-react"
import { cn } from "@/lib/utils"

export const Navbar = () => {
  const { user, logout } = useAuthStore()
  const [isProfileOpen, setIsProfileOpen] = useState(false)

  // Mock notification counts - these would come from TanStack Query polling
  const unreadNotifications = 3

  const navItems = [
    { icon: Home, label: "Home", href: "/feed" },
    { icon: Users, label: "My Network", href: "/networks" },
    { icon: Briefcase, label: "Jobs", href: "/jobs" },
    { icon: MessageSquare, label: "Messaging", href: "/messaging" },
    {
      icon: Bell,
      label: "Notifications",
      href: "/notifications",
      badge: unreadNotifications,
    },
  ]

  return (
    <nav className="sticky top-0 z-50 w-full border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
        {/* Left: Logo & Search */}
        <div className="flex flex-1 items-center gap-2">
          <div className="rounded bg-[#0a66c2] p-1 text-xl leading-none font-bold text-white">
            in
          </div>
          <div className="relative hidden w-full max-w-xs md:block">
            <Search className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-gray-500" />
            <input
              type="text"
              placeholder="Search"
              className="w-full rounded bg-[#edf3f8] py-1.5 pr-4 pl-10 text-sm transition-all focus:ring-1 focus:ring-black focus:outline-none"
            />
          </div>
        </div>

        {/* Right: Navigation Links */}
        <div className="flex h-full items-center gap-1 md:gap-6">
          {navItems.map((item) => (
            <Link
              key={item.label}
              href={item.href}
              className="group relative flex min-w-[64px] flex-col items-center justify-center text-gray-500 transition-colors hover:text-black"
            >
              <div className="relative">
                <item.icon className="h-6 w-6" />
                {item.badge && item.badge > 0 && (
                  <span className="absolute -top-1 -right-1 rounded-full border-2 border-white bg-red-600 px-1 text-[10px] font-bold text-white">
                    {item.badge}
                  </span>
                )}
              </div>
              <span className="mt-1 hidden text-xs font-normal md:block">
                {item.label}
              </span>
              <div className="absolute bottom-0 left-0 h-0.5 w-full scale-x-0 bg-black transition-transform group-hover:scale-x-100" />
            </Link>
          ))}

          {/* Profile Dropdown */}
          <div className="relative ml-2 border-l border-gray-200 pl-4">
            <button
              onClick={() => setIsProfileOpen(!isProfileOpen)}
              className="group flex flex-col items-center"
            >
              <img
                src={user?.avatarUrl || "https://via.placeholder.com/24"}
                alt="Profile"
                className="h-6 w-6 rounded-full border border-gray-200"
              />
              <div className="flex items-center text-gray-500 transition-colors group-hover:text-black">
                <span className="mt-1 hidden text-xs md:block">Me</span>
                <ChevronDown className="mt-1 h-3 w-3" />
              </div>
            </button>

            {/* Dropdown Menu */}
            {isProfileOpen && (
              <div className="absolute right-0 mt-2 w-64 overflow-hidden rounded-lg border border-gray-200 bg-white py-2 shadow-xl">
                <div className="flex items-center gap-3 border-b border-gray-100 px-4 py-2">
                  <img
                    src={user?.avatarUrl}
                    className="h-12 w-12 rounded-full"
                    alt=""
                  />
                  <div>
                    <p className="text-sm font-semibold">{user?.name}</p>
                    <p className="truncate text-xs text-gray-500">
                      {user?.email}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => logout()}
                  className="flex w-full items-center gap-2 px-4 py-3 text-left text-sm text-gray-600 hover:bg-gray-50"
                >
                  <LogOut className="h-4 w-4" />
                  Sign Out
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  )
}
