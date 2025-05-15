"use client"

import { useEffect, useState } from "react"
import { useParams, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { MainNav } from "@/components/main-nav"
import { useToast } from "@/components/ui/use-toast"
import { Loader2 } from "lucide-react"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { useForm } from "react-hook-form"
import { z } from "zod"
import { zodResolver } from "@hookform/resolvers/zod"

interface Answer {
  id: number
  text: string
  correct?: boolean
}

interface Question {
  id: number
  question_text: string
  is_test: boolean
  is_required: boolean
  multipleAnswers: boolean
  answers: Answer[]
}

export default function SurveyPage() {
  const params = useParams()
  const router = useRouter()
  const { toast } = useToast()
  const [questions, setQuestions] = useState<Question[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [formSchema, setFormSchema] = useState<z.ZodObject<any> | null>(null)
  const surveyId = params.id as string

  const form = useForm<any>({
    resolver: formSchema ? zodResolver(formSchema) : undefined,
    mode: "onChange",
    defaultValues: {},
  })

  useEffect(() => {
    const token = localStorage.getItem("token")
    if (!token) {
      router.push("/login")
      toast({
        title: "Требуется авторизация",
        description: "Для прохождения опроса необходимо войти в систему.",
        variant: "destructive",
      })
      return
    }

    const fetchQuestions = async () => {
      setIsLoading(true)
      try {
        const response = await fetch(`http://localhost:8080/surveys/${surveyId}/questions`, {
          headers: { Authorization: `Bearer ${token}` },
        })

        if (!response.ok) throw new Error(`HTTP ошибка ${response.status}`)
        
        const data = await response.json()
        
        const processedQuestions = (data.questions || []).map((q: any) => ({
          id: q.id,
          question_text: q.question_text,
          is_test: Boolean(q.is_test),
          is_required: Boolean(q.is_required),
          multipleAnswers: Boolean(q.multipleAnswers),
          answers: q.is_test
            ? (q.answers || []).map((a: any, index: number) => ({
                id: a.id || index + 1,
                text: a.text || `Ответ ${index + 1}`,
                correct: Boolean(a.correct),
              }))
            : [],
        }))

        setQuestions(processedQuestions)
      } catch (error) {
        console.error("Ошибка загрузки вопросов:", error)
        toast({
          title: "Ошибка загрузки",
          description: "Не удалось загрузить вопросы опроса.",
          variant: "destructive",
        })
        setQuestions([])
      } finally {
        setIsLoading(false)
      }
    }

    fetchQuestions()
  }, [surveyId, router, toast])

  useEffect(() => {
    if (questions.length > 0) {
      const schemaFields: Record<string, z.ZodType<any>> = {}
      const defaultValues: Record<string, any> = {}

      questions.forEach((q) => {
        if (q.is_test) {
          if (q.multipleAnswers) {
            schemaFields[`question_${q.id}`] = q.is_required
              ? z.array(z.string()).min(1, "Выберите хотя бы один ответ")
              : z.array(z.string()).optional()
            defaultValues[`question_${q.id}`] = []
          } else {
            schemaFields[`question_${q.id}`] = q.is_required
              ? z.string().min(1, "Выберите ответ")
              : z.string().optional()
            defaultValues[`question_${q.id}`] = ""
          }
        } else {
          schemaFields[`question_${q.id}`] = q.is_required
            ? z.string().min(1, "Введите ответ")
            : z.string().optional()
          defaultValues[`question_${q.id}`] = ""
        }
      })

      setFormSchema(z.object(schemaFields))
      form.reset(defaultValues)
    }
  }, [questions, form])

  const onSubmit = async (data: any) => {
    setIsSubmitting(true)
    try {
      const token = localStorage.getItem("token")
      if (!token) throw new Error("Токен авторизации не найден")

      const answers: any[] = []
      questions.forEach((question) => {
        const formKey = `question_${question.id}`
        const value = data[formKey] || (question.multipleAnswers ? [] : "")

        if (question.is_test && question.multipleAnswers) {
          value.forEach((answerId: string) => {
            answers.push({
              question_id: question.id,
              answer_id: parseInt(answerId, 10),
            })
          })
        } else if (question.is_test) {
          answers.push({
            question_id: question.id,
            answer_id: parseInt(value, 10),
          })
        } else {
          answers.push({
            question_id: question.id,
            answer_user: value,
          })
        }
      })

      console.log("Структура отправляемых данных:", {
        answers: answers.map(answer => ({
          type: answer.answer_user ? "текстовый ответ" : "тестовый ответ",
          question_id: answer.question_id,
          data: answer.answer_user || { answer_id: answer.answer_id },
          required: questions.find(q => q.id === answer.question_id)?.is_required
        }))
      })

      console.log("Raw JSON payload:", JSON.stringify(
        { answers }, 
        (key, value) => key === 'correct' ? undefined : value, // Исключаем поле correct
        2
      ))

      const response = await fetch("http://localhost:8080/option", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ answers }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || "Ошибка при отправке ответов")
      }

      toast({
        title: "Успешно",
        description: "Ответы отправлены.",
      })

      router.push(`/survey/${surveyId}/results`)
    } catch (error) {
      console.error("Ошибка отправки ответов:", error)
      toast({
        title: "Ошибка",
        description: error instanceof Error ? error.message : "Не удалось отправить ответы",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
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
            <CardTitle className="text-2xl">Прохождение опроса</CardTitle>
            <CardDescription>Ответьте на все вопросы опроса</CardDescription>
          </CardHeader>
          <CardContent>
            {questions.length === 0 ? (
              <p className="text-center text-muted-foreground">В опросе нет вопросов</p>
            ) : (
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                  {questions.map((question) => (
                    <FormField
                      key={question.id}
                      control={form.control}
                      name={`question_${question.id}`}
                      render={({ field }) => (
                        <FormItem className="space-y-3">
                          <FormLabel className="text-base font-medium">
                            {question.question_text}
                            {question.is_required && (
                              <span className="text-red-500 ml-1">*</span>
                            )}
                          </FormLabel>
                          <FormControl>
                            {question.is_test ? (
                              question.answers.length > 0 ? (
                                question.multipleAnswers ? (
                                  <div className="space-y-2">
                                    {question.answers.map((answer) => (
                                      <div key={answer.id} className="flex items-center space-x-2">
                                        <input
                                          type="checkbox"
                                          id={`answer-${answer.id}`}
                                          checked={field.value?.includes(answer.id.toString())}
                                          onChange={(e) => {
                                            const currentValue = Array.isArray(field.value) ? field.value : []
                                            const answerId = answer.id.toString()
                                            if (e.target.checked) {
                                              field.onChange([...currentValue, answerId])
                                            } else {
                                              field.onChange(currentValue.filter((id: string) => id !== answerId))
                                            }
                                          }}
                                          disabled={isSubmitting}
                                          className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                        />
                                        <Label htmlFor={`answer-${answer.id}`} className="font-normal">
                                          {answer.text}
                                        </Label>
                                      </div>
                                    ))}
                                  </div>
                                ) : (
                                  <RadioGroup
                                    onValueChange={field.onChange}
                                    value={field.value}
                                    className="space-y-2"
                                    disabled={isSubmitting}
                                  >
                                    {question.answers.map((answer) => (
                                      <div key={answer.id} className="flex items-center space-x-2">
                                        <RadioGroupItem
                                          value={answer.id.toString()}
                                          id={`answer-${answer.id}`}
                                        />
                                        <Label htmlFor={`answer-${answer.id}`} className="font-normal">
                                          {answer.text}
                                        </Label>
                                      </div>
                                    ))}
                                  </RadioGroup>
                                )
                              ) : (
                                <p className="text-sm text-red-500">
                                  Ошибка: нет вариантов ответов для тестового вопроса
                                </p>
                              )
                            ) : (
                              <Input
                                {...field}
                                placeholder="Введите ваш ответ"
                                disabled={isSubmitting}
                              />
                            )}
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  ))}
                  <Button
                    type="submit"
                    className="w-full"
                    disabled={isSubmitting || !form.formState.isValid}
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Отправка...
                      </>
                    ) : (
                      "Отправить ответы"
                    )}
                  </Button>
                </form>
              </Form>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  )
}