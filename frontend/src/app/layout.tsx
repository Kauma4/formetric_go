import type React from "react"
import type { Metadata } from "next"
import { Inter } from 'next/font/google'
import "@/app/globals.css"
import { Toaster } from "@/components/ui/toaster"

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "Survey Application",
  description: "A modern survey application",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="ru">
      <body className={inter.className}>
        {children}
        <Toaster />
      </body>
    </html>
  )
}