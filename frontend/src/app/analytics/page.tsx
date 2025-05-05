"use client"

import { useEffect, useState } from "react"
import { CombinedChart } from "@/components/charts"
import { MainNav } from "@/components/main-nav"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useToast } from "@/components/ui/use-toast"
import { Loader2 } from "lucide-react"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

interface SurveyStats {
  id: number
  title: string
  responseCount: number
}

interface AnalyticsData {
  totalSurveys: number
  totalResponses: number
  popularSurveys: SurveyStats[]
  responseRate: number
}

export default function AnalyticsPage() {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [timeRange, setTimeRange] = useState("all")
  const { toast } = useToast()

  useEffect(() => {
    // Добавляем фиктивные данные для красоты
    setIsLoading(true)
    const fakeData: AnalyticsData = {
      totalSurveys: 120,
      totalResponses: 980,
      responseRate: 82,
      popularSurveys: [
        { id: 1, title: "Опрос о качестве сервиса", responseCount: 450 },
        { id: 2, title: "Опрос о продукте", responseCount: 300 },
        { id: 3, title: "Опрос о цене", responseCount: 150 },
        { id: 4, title: "Опрос о доставке", responseCount: 80 },
        { id: 5, title: "Опрос о поддержке", responseCount: 60 },
      ],
    }
    setData(fakeData)
    setIsLoading(false)
  }, [])

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-6 space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">Аналитика опросов</h1>
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Период" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="week">За неделю</SelectItem>
              <SelectItem value="month">За месяц</SelectItem>
              <SelectItem value="year">За год</SelectItem>
              <SelectItem value="all">За все время</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {isLoading ? (
          <div className="flex justify-center">
            <Loader2 className="h-8 w-8 animate-spin" />
          </div>
        ) : data && (
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Общая статистика</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span>Всего опросов:</span>
                  <span className="font-medium">{data.totalSurveys}</span>
                </div>
                <div className="flex justify-between">
                  <span>Всего ответов:</span>
                  <span className="font-medium">{data.totalResponses}</span>
                </div>
                <div className="flex justify-between">
                  <span>Процент ответов:</span>
                  <span className="font-medium">{data.responseRate}%</span>
                </div>
              </CardContent>
            </Card>

            {/* Графики */}
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>Графики</CardTitle>
              </CardHeader>
              <CardContent>
                <CombinedChart
                  data={data.popularSurveys.map(item => ({
                    name: item.title,
                    value: item.responseCount,
                  }))}
                />
              </CardContent>
            </Card>
          </div>
        )}
      </main>
    </div>
  )
}
