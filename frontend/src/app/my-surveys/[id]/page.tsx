"use client"

import { useEffect, useState } from "react"
import { useParams, useRouter } from "next/navigation"
import { MainNav } from "@/components/main-nav"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Loader2 } from "lucide-react"
import { useToast } from "@/components/ui/use-toast"
import { format } from "date-fns"
import { ru } from "date-fns/locale"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

interface Participant {
  user_id: number
  username: string
  total_score: number
  date: string
  max_score: number
}

interface UserAnswer {
  question_text: string
  user_answer: string
  correct_answer: string
  is_correct?: boolean
}

export default function SurveyParticipants() {
  const { id } = useParams<{ id: string }>()
  const router = useRouter()
  const { toast } = useToast()
  const [participants, setParticipants] = useState<Participant[]>([])
  const [selectedUser, setSelectedUser] = useState<number | null>(null)
  const [userAnswers, setUserAnswers] = useState<UserAnswer[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isLoadingAnswers, setIsLoadingAnswers] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  useEffect(() => {
    const fetchParticipants = async () => {
      try {
        const token = localStorage.getItem("token")
        if (!token) {
          toast({
            title: "Ошибка",
            description: "Требуется авторизация",
            variant: "destructive",
          })
          router.push("/login")
          return
        }

        const response = await fetch(`http://localhost:8080/surveys/${id}/participants`, {
          headers: { Authorization: `Bearer ${token}` }
        })

        if (!response.ok) {
          throw new Error("Ошибка загрузки участников")
        }

        const data = await response.json()
        setParticipants(data.participants || [])
      } catch (error) {
        toast({
          title: "Ошибка",
          description: error instanceof Error ? error.message : "Неизвестная ошибка",
          variant: "destructive",
        })
        setParticipants([])
      } finally {
        setIsLoading(false)
      }
    }

    fetchParticipants()
  }, [id, router, toast])

  const fetchUserAnswers = async (userId: number) => {
    setIsLoadingAnswers(true)
    try {
      const token = localStorage.getItem("token")
      if (!token) throw new Error("Требуется авторизация")

      const response = await fetch(
        `http://localhost:8080/survey/${id}/answers/user`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ user_id: userId }),
        }
      )

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || "Ошибка загрузки ответов")
      }

      const data = await response.json()
      setUserAnswers(data.answers || [])
    } catch (error) {
      toast({
        title: "Ошибка",
        description: error instanceof Error ? error.message : "Неизвестная ошибка",
        variant: "destructive",
      })
      setUserAnswers([])
    } finally {
      setIsLoadingAnswers(false)
    }
  }

  const handleUserClick = async (userId: number) => {
    setSelectedUser(userId)
    setIsDialogOpen(true)
    await fetchUserAnswers(userId)
  }

  const formatDate = (dateString: string) => {
    try {
      return format(new Date(dateString), "dd MMM yyyy, HH:mm", { locale: ru })
    } catch {
      return "Неверная дата"
    }
  }

  const calculatePercentage = (score: number, max: number) => {
    if (max <= 0) return "0.0"
    return ((score / max) * 100).toFixed(1)
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-8">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold">Участники опроса</h1>
          <Button onClick={() => router.back()}>Назад к опросу</Button>
        </div>

        {isLoading ? (
          <div className="flex justify-center">
            <Loader2 className="h-8 w-8 animate-spin" />
          </div>
        ) : participants.length === 0 ? (
          <div className="text-center text-muted-foreground">
            Пока никто не прошел этот опрос
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Пользователь</TableHead>
                  <TableHead>Дата прохождения</TableHead>
                  <TableHead>Баллы</TableHead>
                  <TableHead>Максимум</TableHead>
                  <TableHead>Процент</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {participants.map((participant) => (
                  <TableRow 
                    key={participant.user_id}
                    className="cursor-pointer hover:bg-gray-50"
                    onClick={() => handleUserClick(participant.user_id)}
                  >
                    <TableCell>{participant.username}</TableCell>
                    <TableCell>{formatDate(participant.date)}</TableCell>
                    <TableCell>{participant.total_score}</TableCell>
                    <TableCell>{participant.max_score}</TableCell>
                    <TableCell>
                      {calculatePercentage(participant.total_score, participant.max_score)}%
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>
                Ответы пользователя {participants.find(p => p.user_id === selectedUser)?.username || ""}
              </DialogTitle>
            </DialogHeader>
            <div className="max-h-[60vh] overflow-y-auto">
              {isLoadingAnswers ? (
                <div className="flex justify-center py-4">
                  <Loader2 className="h-6 w-6 animate-spin" />
                </div>
              ) : userAnswers.length === 0 ? (
                <div className="text-center text-muted-foreground py-4">
                  Нет ответов для отображения
                </div>
              ) : (
                <div className="space-y-6">
                  {userAnswers.map((answer, index) => (
                    <div key={index} className="border-b pb-4">
                      <div className="font-medium mb-2">{answer.question_text}</div>
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">Ответ:</span>
                          <span className="font-medium">
                            {answer.user_answer || "Нет ответа"}
                          </span>
                        </div>
                        
                        {answer.correct_answer && (
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Правильный ответ:</span>
                            <span className="text-green-600">
                              {answer.correct_answer}
                            </span>
                          </div>
                        )}
                        
                        {typeof answer.is_correct !== 'undefined' && (
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Результат:</span>
                            <span className={answer.is_correct ? "text-green-600" : "text-red-600"}>
                              {answer.is_correct ? "Правильно" : "Неправильно"}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
            <div className="flex justify-end mt-4">
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Закрыть
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </main>
    </div>
  )
}