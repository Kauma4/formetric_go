import { NextResponse } from "next/server"

interface Message {
  content: string
  data?: any
}

export async function POST(req: Request) {
  try {
    const { messages }: { messages?: Message[] } = await req.json()

    if (!messages || messages.length === 0) {
      return NextResponse.json({ error: "No messages provided" }, { status: 400 })
    }

    const lastMessage = messages[messages.length - 1]

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

    const path = lastMessage.content.startsWith('/') ? lastMessage.content : `/${lastMessage.content}`
    const response = await fetch(`${apiUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(lastMessage.data || {}),
    })

    if (!response.ok) {
      throw new Error(`HTTP error ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error("Error:", error)
    return NextResponse.json({ error: "An error occurred while processing your request" }, { status: 500 })
  }
}