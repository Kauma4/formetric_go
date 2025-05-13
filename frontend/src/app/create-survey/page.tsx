"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { MainNav } from "@/components/main-nav"
import { useToast } from "@/components/ui/use-toast"
import { Loader2, Plus, Trash2 } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"

interface AnswerForm {
  text: string
  isCorrect: boolean
}

interface QuestionForm {
  text: string
  ball: number
  type: 'test' | 'text' | 'free_text'
  required: boolean
  multipleAnswers: boolean
  answers: AnswerForm[]
  hasCorrectAnswer: boolean
}

export default function CreateSurveyPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [title, setTitle] = useState("")
  const [description, setDescription] = useState("")
  const [isPrivate, setIsPrivate] = useState(false)
  const [questions, setQuestions] = useState<QuestionForm[]>([
    { 
      text: "", 
      ball: 0, 
      type: "test",
      required: true,
      multipleAnswers: false,
      hasCorrectAnswer: true,
      answers: [{ text: "", isCorrect: true }] 
    },
  ])
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem("token")
    if (!token) {
      toast({
        title: "Требуется авторизация",
        description: "Для создания опроса необходимо войти в систему.",
        variant: "destructive",
      })
      router.push("/login")
    }
  }, [router, toast])

  const handleAddQuestion = () => {
    setQuestions([
      ...questions, 
      { 
        text: "", 
        ball: 1, 
        type: "test",
        required: true,
        multipleAnswers: false,
        hasCorrectAnswer: true,
        answers: [{ text: "", isCorrect: true }] 
      }
    ])
  }

  const handleRemoveQuestion = (index: number) => {
    if (questions.length > 1) {
      const newQuestions = [...questions]
      newQuestions.splice(index, 1)
      setQuestions(newQuestions)
    }
  }

  const handleQuestionChange = (index: number, field: string, value: any) => {
    const updatedQuestions = [...questions]
    
    if (field === 'type') {
      updatedQuestions[index].hasCorrectAnswer = value !== 'free_text'
      updatedQuestions[index].answers = value === 'test' 
        ? [{ text: "", isCorrect: true }] 
        : value === 'text'
        ? [{ text: "", isCorrect: false }]
        : []
      updatedQuestions[index].multipleAnswers = false
    }
    
    updatedQuestions[index] = { ...updatedQuestions[index], [field]: value }
    setQuestions(updatedQuestions)
  }

  const handleAddAnswer = (qIndex: number) => {
    const updatedQuestions = [...questions]
    updatedQuestions[qIndex].answers.push({ text: "", isCorrect: false })
    setQuestions(updatedQuestions)
  }

  const handleRemoveAnswer = (qIndex: number, aIndex: number) => {
    if (questions[qIndex].answers.length > 1) {
      const updatedQuestions = [...questions]
      updatedQuestions[qIndex].answers.splice(aIndex, 1)

      if (!updatedQuestions[qIndex].multipleAnswers && 
          updatedQuestions[qIndex].answers.every((a) => !a.isCorrect)) {
        updatedQuestions[qIndex].answers[0].isCorrect = true
      }

      setQuestions(updatedQuestions)
    }
  }

  const handleAnswerChange = (qIndex: number, aIndex: number, value: string) => {
    const updatedQuestions = [...questions]
    updatedQuestions[qIndex].answers[aIndex].text = value
    setQuestions(updatedQuestions)
  }

  const handleSetCorrectAnswer = (qIndex: number, aIndex: number) => {
    const updatedQuestions = [...questions]
    
    if (updatedQuestions[qIndex].multipleAnswers) {
      updatedQuestions[qIndex].answers[aIndex].isCorrect = 
        !updatedQuestions[qIndex].answers[aIndex].isCorrect
    } else {
      updatedQuestions[qIndex].answers = updatedQuestions[qIndex].answers.map((ans, i) => ({
        ...ans,
        isCorrect: i === aIndex,
      }))
    }
    
    setQuestions(updatedQuestions)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!title.trim()) {
      toast({
        title: "Ошибка валидации",
        description: "Введите название опроса",
        variant: "destructive",
      })
      return
    }

    for (let i = 0; i < questions.length; i++) {
      if (!questions[i].text.trim()) {
        toast({
          title: "Ошибка валидации",
          description: `Введите текст вопроса ${i + 1}`,
          variant: "destructive",
        })
        return
      }

      if (questions[i].type === 'test') {
        for (let j = 0; j < questions[i].answers.length; j++) {
          if (!questions[i].answers[j].text.trim()) {
            toast({
              title: "Ошибка валидации",
              description: `Введите текст ответа ${j + 1} для вопроса ${i + 1}`,
              variant: "destructive",
            })
            return
          }
        }
        
        if (!questions[i].answers.some(a => a.isCorrect)) {
          toast({
            title: "Ошибка валидации",
            description: `Выберите хотя бы один правильный ответ для вопроса ${i + 1}`,
            variant: "destructive",
          })
          return
        }
      } else if (questions[i].type === 'text') {
        if (questions[i].hasCorrectAnswer && !questions[i].answers[0]?.text.trim()) {
          toast({
            title: "Ошибка валидации",
            description: `Введите правильный ответ для текстового вопроса ${i + 1}`,
            variant: "destructive",
          })
          return
        }
      }
    }

    setIsSubmitting(true)
    const token = localStorage.getItem("token")

    try {
      const surveyResponse = await fetch("http://localhost:8080/survey", {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          title,
          description,
          is_private: isPrivate
        })
      })

      if (!surveyResponse.ok) {
        const errorData = await surveyResponse.json()
        throw new Error(errorData.error || "Ошибка создания опроса")
      }

      const surveyData = await surveyResponse.json()
      const surveyId = surveyData.survey_id

      for (const question of questions) {
    const questionData: any = {
  survey_id: surveyId,
  question_text: question.text,
  ball: question.ball,
  required: question.required,
  is_test: question.type === 'test',
  multiple_answers: question.multipleAnswers,
  answers: [],
  correct_answer: "",
}

if (question.type === 'test') {
  questionData.answers = question.answers.map(a => ({
    text: a.text,
    correct: a.isCorrect
  }))
}

if (question.type === 'text') {
  questionData.correct_answer = question.answers[0]?.text || ""
}

if (question.type === 'free_text') {
  questionData.correct_answer = ""
}


        const questionResponse = await fetch("http://localhost:8080/question", {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json"
          },
          body: JSON.stringify(questionData)
        })

        if (!questionResponse.ok) {
          const errorData = await questionResponse.json()
          throw new Error(errorData.error || `Ошибка создания вопроса "${question.text}"`)
        }
      }

      toast({
        title: "Успешно",
        description: "Опрос успешно создан",
      })

      router.push("/surveys")
    } catch (error) {
      console.error("Ошибка создания опроса:", error)
      toast({
        title: "Ошибка создания",
        description: error instanceof Error ? error.message : "Не удалось создать опрос",
        variant: "destructive",
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl">Создание нового опроса</CardTitle>
            <CardDescription>Заполните форму для создания опроса</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-8">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="title">Название опроса</Label>
                  <Input
                    id="title"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    placeholder="Введите название опроса"
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Описание</Label>
                  <Textarea
                    id="description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Введите описание опроса"
                    rows={3}
                  />
                </div>
              </div>
              
              <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setIsPrivate(!isPrivate)}
                    className={`relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                      isPrivate ? 'bg-primary' : 'bg-muted'
                    }`}
                  >
                    <span
                      className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-lg ring-0 transition duration-200 ease-in-out ${
                        isPrivate ? 'translate-x-5' : 'translate-x-0'
                      }`}
                    />
                  </button>
                  <Label htmlFor="privacy">
                    Приватный опрос
                  </Label>
                </div>
                <p className="text-sm text-muted-foreground">
                  {isPrivate 
                    ? 'Доступен только по прямой ссылке' 
                    : 'Будет отображаться в общем списке'}
                </p>
              </div>

              <Separator />

              <div className="space-y-6 relative pb-20">
                {questions.map((question, qIndex) => (
                  <Card key={qIndex} className="border border-muted">
                    <CardHeader className="p-4">
                      <div className="flex items-start justify-between">
                        <div className="space-y-1 w-full">
                          <Label htmlFor={`question-${qIndex}`}>Вопрос {qIndex + 1}</Label>
                          <Input
                            id={`question-${qIndex}`}
                            value={question.text}
                            onChange={(e) => handleQuestionChange(qIndex, "text", e.target.value)}
                            placeholder="Введите текст вопроса"
                            required
                          />

                          <div className="flex items-center space-x-4 mt-2 flex-wrap gap-y-2">
                            <div className="flex items-center space-x-2">
                              <Label htmlFor={`type-${qIndex}`}>Тип вопроса:</Label>
                              <select
                                id={`type-${qIndex}`}
                                value={question.type}
                                onChange={(e) => handleQuestionChange(qIndex, "type", e.target.value)}
                                className="rounded-md border border-input bg-background px-3 py-2 text-sm"
                              >
                                <option value="test">Тестовый (выбор варианта)</option>
                                <option value="text">Текстовый с проверкой</option>
                                <option value="free_text">Свободный ответ</option>
                              </select>
                            </div>
                            
                            <div className="flex items-center space-x-2">
                              <Label htmlFor={`required-${qIndex}`}>Обязательность:</Label>
                              <select
                                id={`required-${qIndex}`}
                                value={question.required ? "required" : "optional"}
                                onChange={(e) => handleQuestionChange(qIndex, "required", e.target.value === "required")}
                                className="rounded-md border border-input bg-background px-3 py-2 text-sm"
                              >
                                <option value="required">Обязательный</option>
                                <option value="optional">Необязательный</option>
                              </select>
                            </div>
                            
                            {question.type === 'test' && (
                              <div className="flex items-center space-x-2">
                                <Label htmlFor={`multiple-${qIndex}`}>Несколько правильных:</Label>
                                <input
                                  id={`multiple-${qIndex}`}
                                  type="checkbox"
                                  checked={question.multipleAnswers}
                                  onChange={(e) => handleQuestionChange(qIndex, "multipleAnswers", e.target.checked)}
                                  className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                />
                              </div>
                            )}
                            
                            <div className="flex items-center space-x-2">
                              <Label htmlFor={`ball-${qIndex}`} className="text-sm">
                                Баллы:
                              </Label>
                              <Input
                                id={`ball-${qIndex}`}
                                type="number"
                                value={question.ball}
                                onChange={(e) => handleQuestionChange(qIndex, "ball", Number(e.target.value))}
                                className="w-16"
                                min="0"
                              />
                            </div>
                          </div>
                        </div>
                        
                        {questions.length > 1 && (
                          <Button
                            type="button"
                            onClick={() => handleRemoveQuestion(qIndex)}
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-destructive ml-2"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </CardHeader>
                    <CardContent className="p-4 pt-0">
                      {question.type === 'test' ? (
                        <div className="space-y-4">
                          <div className="flex items-center justify-between">
                            <Label>Варианты ответов</Label>
                            <Button type="button" onClick={() => handleAddAnswer(qIndex)} variant="outline" size="sm">
                              <Plus className="h-4 w-4 mr-2" />
                              Добавить ответ
                            </Button>
                          </div>

                          <div className="space-y-3">
                            {question.answers.map((answer, aIndex) => (
                              <div key={aIndex} className="space-y-2">
                                <div className="flex items-center space-x-2">
                                  {question.multipleAnswers ? (
                                    <input
                                      type="checkbox"
                                      checked={answer.isCorrect}
                                      onChange={() => handleSetCorrectAnswer(qIndex, aIndex)}
                                      className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                    />
                                  ) : (
                                    <input
                                      type="radio"
                                      checked={answer.isCorrect}
                                      onChange={() => handleSetCorrectAnswer(qIndex, aIndex)}
                                      className="h-4 w-4 text-indigo-600 focus:ring-indigo-500"
                                    />
                                  )}
                                  <Input
                                    value={answer.text}
                                    onChange={(e) => handleAnswerChange(qIndex, aIndex, e.target.value)}
                                    placeholder={`Вариант ответа ${aIndex + 1}`}
                                    className="flex-1"
                                    required
                                  />
                                  {question.answers.length > 1 && (
                                    <Button
                                      type="button"
                                      onClick={() => handleRemoveAnswer(qIndex, aIndex)}
                                      variant="ghost"
                                      size="icon"
                                      className="h-8 w-8 text-destructive"
                                    >
                                      <Trash2 className="h-4 w-4" />
                                    </Button>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      ) : question.type === 'text' ? (
                        <div className="space-y-4">
                          <div className="space-y-2">
                            <Label>Правильный ответ (для проверки)</Label>
                            <Input
                              value={question.answers[0]?.text || ''}
                              onChange={(e) => handleAnswerChange(qIndex, 0, e.target.value)}
                              placeholder="Введите правильный ответ"
                            />
                          </div>
                        </div>
                      ) : (
                        <div className="space-y-4">
                          <div className="space-y-2">
                            <Label>Поле для свободного ответа</Label>
                            <Input
                              disabled
                              placeholder="Пользователь введет свой ответ здесь"
                              className="bg-muted"
                            />
                          </div>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                ))}

                {/* Плавающая кнопка добавления вопроса */}
                <div className="fixed bottom-6 right-6 z-50">
                  <Button 
                    type="button" 
                    onClick={handleAddQuestion} 
                    size="lg"
                    className="rounded-full shadow-lg h-14 w-14 p-0 bg-primary hover:bg-primary/90 transition-colors"
                  >
                    <Plus className="h-6 w-6" />
                  </Button>
                </div>
              </div>

              <Button type="submit" className="w-full" disabled={isSubmitting}>
                {isSubmitting ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Создание...
                  </>
                ) : (
                  "Создать опрос"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>
      </main>
    </div>
  )
}