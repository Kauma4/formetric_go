"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { MainNav } from "@/components/main-nav"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Loader2 } from "lucide-react"

interface Survey {
  id: number
  title: string
  description: string
  created_at: string
}

export default function MySurveys() {
  const router = useRouter()
  const [surveys, setSurveys] = useState<Survey[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchSurveys = async () => {
      try {
        const token = localStorage.getItem("token")
        if (!token) {
          router.push("/login")
          return
        }

        const response = await fetch("http://localhost:8080/my-surveys", {
          headers: { Authorization: `Bearer ${token}` }
        })

        const data = await response.json()
        setSurveys(data.surveys)
      } finally {
        setIsLoading(false)
      }
    }

    fetchSurveys()
  }, [router])

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Мои опросы</h1>
      <div className="grid gap-4 md:grid-cols-2">
        {surveys.map(survey => (
          <Card 
            key={survey.id}
            className="cursor-pointer hover:bg-gray-50"
            onClick={() => router.push(`/my-surveys/${survey.id}`)}
          >
            <CardHeader>
              <CardTitle>{survey.title}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-600">{survey.description}</p>
              <p className="text-sm text-gray-500 mt-2">
                Создан: {new Date(survey.created_at).toLocaleDateString()}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}