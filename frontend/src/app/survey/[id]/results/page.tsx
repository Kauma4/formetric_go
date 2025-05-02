"use client"

import { useEffect, useState } from "react"
import { useParams, useRouter } from "next/navigation"
import { MainNav } from "@/components/main-nav"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Loader2, CheckCircle2, XCircle } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { ExclamationTriangleIcon } from "@radix-ui/react-icons"

interface Result {
  total_score: number
  max_score: number
}

interface UserAnswer {
  question_text: string
  user_answer: string | number
  is_correct?: boolean
  correct_answer?: string | number
  question_type?: 'test' | 'text'
}

export default function SurveyResults() {
  const params = useParams()
  const router = useRouter()
  const { toast } = useToast()
  const [results, setResults] = useState<Result | null>(null)
  const [userAnswers, setUserAnswers] = useState<UserAnswer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const surveyId = parseInt(params.id as string, 10)

  const fetchData = async () => {
    setIsLoading(true)
    setError(null)
    
    try {
      const token = localStorage.getItem("token")
      if (!token) {
        router.push("/login")
        throw new Error("Требуется авторизация")
      }

      if (isNaN(surveyId)) {
        throw new Error("Неверный ID опроса")
      }

      const [resultsRes, answersRes] = await Promise.all([
        fetch(`http://localhost:8080/survey/${surveyId}/results`, {
          headers: { 
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json"
          }
        }),
        fetch(`http://localhost:8080/survey/${surveyId}/answers/simple`, {
          headers: { 
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json"
          }
        })
      ])

      if (!resultsRes.ok) {
        const errorData = await resultsRes.json()
        throw new Error(errorData.error || "Ошибка загрузки результатов")
      }

      if (!answersRes.ok) {
        const errorData = await answersRes.json()
        throw new Error(errorData.error || "Ошибка загрузки ответов")
      }

      const resultsData = await resultsRes.json()
      const answersData = await answersRes.json()

      // Валидация структуры ответа
      if (typeof resultsData.total_score !== 'number' || 
          typeof resultsData.max_score !== 'number') {
        throw new Error("Неверный формат результатов")
      }

      if (!Array.isArray(answersData.answers)) {
        throw new Error("Неверный формат ответов")
      }

      setResults({
        total_score: resultsData.total_score,
        max_score: resultsData.max_score
      })

      setUserAnswers(answersData.answers.map((a: any) => ({
        question_text: a.question_text,
        user_answer: a.user_answer,
        is_correct: a.is_correct,
        correct_answer: a.correct_answer,
        question_type: a.is_correct !== undefined ? 'test' : 'text'
      })))

    } catch (error) {
      console.error("Ошибка:", error)
      setError(error instanceof Error ? error.message : "Неизвестная ошибка")
      
      toast({
        title: "Ошибка загрузки",
        description: "Не удалось получить данные результатов",
        variant: "destructive",
      })

    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [surveyId])

  const handleRetry = () => {
    fetchData()
  }

  if (error) {
    return (
      <div className="flex min-h-screen flex-col">
        <MainNav />
        <main className="flex-1 container py-6">
          <Card className="max-w-3xl mx-auto">
            <CardContent className="pt-6">
              <Alert variant="destructive">
                <ExclamationTriangleIcon className="h-4 w-4" />
                <AlertTitle>Ошибка</AlertTitle>
                <AlertDescription className="space-y-4">
                  <div>{error}</div>
                  <Button 
                    size="sm" 
                    onClick={handleRetry}
                  >
                    Повторить попытку
                  </Button>
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        </main>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen flex-col">
        <MainNav />
        <main className="flex-1 container py-6 flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </main>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-6">
        <Card className="max-w-3xl mx-auto">
          <CardHeader>
            <CardTitle className="text-2xl">Результаты опроса</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {results && (
              <div className="text-center">
                <h3 className="text-3xl font-bold">
                  {results.total_score} из {results.max_score} баллов
                </h3>
                <p className="text-muted-foreground mt-2">
                  Правильных ответов: {Math.round((results.total_score / results.max_score) * 100)}%
                </p>
              </div>
            )}

            <div className="space-y-4">
              {userAnswers.map((answer, index) => (
                <Card 
                  key={index}
                  className={`p-4 ${
                    answer.is_correct === true 
                      ? 'bg-green-50 border-green-200' 
                      : answer.is_correct === false 
                        ? 'bg-red-50 border-red-200' 
                        : 'bg-gray-50 border-gray-200'
                  }`}
                >
                  <h4 className="font-medium">{answer.question_text}</h4>
                  <div className="mt-2 space-y-2">
                    <p className="text-sm">
                      Ваш ответ:{" "}
                      <span className={
                        answer.is_correct === true 
                          ? "text-green-600" 
                          : answer.is_correct === false 
                            ? "text-red-600" 
                            : "text-gray-600"
                      }>
                        {answer.user_answer}
                      </span>
                    </p>
                    
                    {(answer.question_type === 'test' || answer.correct_answer) && (
                      <p className="text-sm text-muted-foreground">
                        Правильный ответ: {answer.correct_answer}
                      </p>
                    )}
                    
                    {answer.question_type === 'test' && answer.is_correct !== undefined && (
                      <div className="flex items-center gap-2 mt-1">
                        {answer.is_correct ? (
                          <CheckCircle2 className="h-4 w-4 text-green-600" />
                        ) : (
                          <XCircle className="h-4 w-4 text-red-600" />
                        )}
                        <span className={answer.is_correct ? "text-green-600" : "text-red-600"}>
                          {answer.is_correct ? "Правильно" : "Неправильно"}
                        </span>
                      </div>
                    )}
                  </div>
                </Card>
              ))}
            </div>

            <Button 
              onClick={() => router.push("/surveys")}
              className="w-full mt-6"
            >
              Вернуться к списку опросов
            </Button>
          </CardContent>
        </Card>
      </main>
    </div>
  )
}