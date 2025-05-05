"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { MainNav } from "@/components/main-nav"
import { useToast } from "@/components/ui/use-toast"
import { Loader2, Plus, Pencil, Trash2 } from "lucide-react"
import Link from "next/link"

interface Survey {
  id: number
  title: string
  description: string
  questionCount: number
  estimatedTime: number
}

export default function SurveysPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [surveys, setSurveys] = useState<Survey[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem("token")
    setIsLoggedIn(!!token)

    const fetchSurveys = async () => {
      setIsLoading(true)
      try {
        if (!token) {
          toast({
            title: "Требуется авторизация",
            description: "Для просмотра опросов необходимо войти в систему.",
            variant: "destructive",
          })
          router.push("/login")
          return
        }

        const response = await fetch("http://localhost:8080/surveys", {
          headers: { Authorization: `Bearer ${token}` },
        })

        if (!response.ok) {
          throw new Error("HTTP error " + response.status)
        }

        const data = await response.json()
        setSurveys(data)
      } catch (error) {
        console.error("Ошибка при получении опросов:", error)
        toast({
          title: "Ошибка загрузки",
          description: "Не удалось загрузить список опросов.",
          variant: "destructive",
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchSurveys()
  }, [router, toast])

  const handleDelete = async (surveyId: number) => {
    if (!confirm("Вы уверены, что хотите удалить этот опрос?")) return
    
    try {
      const token = localStorage.getItem("token")
      const response = await fetch(`http://localhost:8080/survey/${surveyId}`, { 
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || "Ошибка удаления")
      }

      setSurveys(prev => prev.filter(s => s.id !== surveyId))
      toast({
        title: "Опрос удален",
        description: "Опрос был успешно удален",
      })
    } catch (error) {
      console.error("Ошибка при удалении:", error)
      toast({
        title: "Ошибка удаления",
        description: error instanceof Error ? error.message : "Не удалось удалить опрос",
        variant: "destructive",
      })
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-6">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">Доступные опросы</h1>
          {isLoggedIn && (
            <Link href="/create-survey">
              <Button>
                <Plus className="mr-2 h-4 w-4" />
                Создать опрос
              </Button>
            </Link>
          )}
        </div>
        {isLoading ? (
          <div className="flex justify-center items-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : (surveys ?? []).length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center h-64">
              <p className="text-muted-foreground mb-4">Опросов пока нет</p>
              {isLoggedIn && (
                <Link href="/create-survey">
                  <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    Создать первый опрос
                  </Button>
                </Link>
              )}
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {surveys.map((survey) => (
              <Card key={survey.id} className="flex flex-col">
                <CardHeader>
                  <CardTitle>{survey.title}</CardTitle>
                  <CardDescription>{survey.description}</CardDescription>
                </CardHeader>
                <CardContent className="flex-1 space-y-2">
                  <div className="flex justify-between text-sm text-muted-foreground">
                    <span>Вопросов: {survey.questionCount}</span>
                    <span>~{survey.estimatedTime} мин</span>
                  </div>
                </CardContent>
                <CardFooter className="flex justify-between gap-2">
                  <Button 
                    variant="outline" 
                    className="flex-1" 
                    onClick={() => router.push(`/survey/${survey.id}`)}
                  >
                    Пройти опрос
                  </Button>
                  {isLoggedIn && (
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => router.push(`/edit-survey/${survey.id}`)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(survey.id)}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  )}
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}