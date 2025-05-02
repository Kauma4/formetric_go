// app/my-surveys/[id]/participants/page.tsx
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

export default function SurveyParticipants() {
  const { id } = useParams()
  const router = useRouter()
  const { toast } = useToast()
  const [participants, setParticipants] = useState<Participant[]>([])
  const [selectedUser, setSelectedUser] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isDialogOpen, setIsDialogOpen] = useState(false)

  useEffect(() => {
    const fetchParticipants = async () => {
      try {
        const token = localStorage.getItem("token")
        if (!token) {
          router.push("/login")
          return
        }

        const response = await fetch(`http://localhost:8080/surveys/${id}/participants`, {
          headers: { Authorization: `Bearer ${token}` }
        })

        const data = await response.json()
        setParticipants(data.participants)
      } catch (error) {
        toast({
          title: "Ошибка",
          description: error instanceof Error ? error.message : "Неизвестная ошибка",
          variant: "destructive",
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchParticipants()
  }, [id, router, toast])

  const formatDate = (dateString: string) => {
    return format(new Date(dateString), "dd MMM yyyy, HH:mm", { locale: ru })
  }

  const calculatePercentage = (score: number, max: number) => {
    return max > 0 ? ((score / max) * 100).toFixed(1) : 0
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
                    onClick={() => {
                      setSelectedUser(participant.user_id)
                      setIsDialogOpen(true)
                    }}
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
              <DialogTitle>Ответы пользователя</DialogTitle>
            </DialogHeader>
            <div className="max-h-[60vh] overflow-y-auto">
              {/* Здесь будет компонент с детальными ответами */}
              {selectedUser && (
                <div className="space-y-4">
                  <p>Детальная информация для пользователя ID: {selectedUser}</p>
                  {/* Добавьте отображение вопросов и ответов */}
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      </main>
    </div>
  )
}