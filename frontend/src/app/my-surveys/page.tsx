"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { MainNav } from "@/components/main-nav"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Loader2 } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"

interface Survey {
  id: number
  title: string
  description: string
  created_at: string
}

export default function MySurveys() {
  const router = useRouter()
  const { toast } = useToast()
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

        if (!response.ok) {
          throw new Error("Ошибка загрузки опросов")
        }

        const data = await response.json()
        setSurveys(Array.isArray(data.surveys) ? data.surveys : [])
      } catch (error) {
        toast({
          title: "Ошибка",
          description: error instanceof Error ? error.message : "Не удалось загрузить опросы",
          variant: "destructive",
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchSurveys()
  }, [router, toast])

  const downloadSurveyPdf = async (surveyId: number) => {
    try {
      const token = localStorage.getItem("token")
      if (!token) throw new Error("Вы не авторизованы")

      const response = await fetch(`http://localhost:8080/surveys/${surveyId}/analytics/pdf`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        throw new Error("Не удалось скачать PDF")
      }

      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)

      const a = document.createElement("a")
      a.href = url
      a.download = `survey_${surveyId}_analytics.pdf`
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      toast({
        title: "Ошибка",
        description: error instanceof Error ? error.message : "Ошибка скачивания PDF",
        variant: "destructive",
      })
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">Мои опросы</h1>
          <Button variant="outline" onClick={() => router.back()}>
            Назад
          </Button>
        </div>

        {surveys.length === 0 ? (
          <div className="text-center text-muted-foreground">
            У вас пока нет созданных опросов
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2">
            {surveys.map(survey => (
              <Card 
                key={survey.id}
                className="group cursor-pointer hover:bg-gray-50"
              >
                <CardHeader onClick={() => router.push(`/my-surveys/${survey.id}`)}>
                  <CardTitle>{survey.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-gray-600">{survey.description}</p>
                  <p className="text-sm text-gray-500 mt-2">
                    Создан: {new Date(survey.created_at).toLocaleDateString()}
                  </p>
                  <Button
                    variant="secondary"
                    className="mt-4"
                    onClick={(e) => {
                      e.stopPropagation()
                      downloadSurveyPdf(survey.id)
                    }}
                  >
                    Скачать PDF
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
