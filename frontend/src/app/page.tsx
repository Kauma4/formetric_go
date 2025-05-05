"use client"

import { useRouter } from "next/navigation"
import { useEffect } from "react"
import { MainNav } from "@/components/main-nav"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import Link from "next/link"

export default function Home() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem("token")
    if (token) {
      router.push("/surveys") // Перенаправляем авторизованных пользователей
    }
  }, [router])

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1">
        <section className="w-full py-12 md:py-24 lg:py-32">
          <div className="container px-4 md:px-6">
            <div className="flex flex-col items-center justify-center space-y-4 text-center">
              <div className="space-y-2">
                <h1 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl lg:text-6xl">
                  Создавайте и проходите опросы
                </h1>
                <p className="mx-auto max-w-[700px] text-gray-500 md:text-xl">
                  Простая и удобная платформа для создания опросов и сбора ответов
                </p>
              </div>
              <div className="flex flex-col gap-2 min-[400px]:flex-row">
                <Link href="/register">
                  <Button size="lg">Регистрация</Button>
                </Link>
                <Link href="/login">
                  <Button size="lg" variant="outline">
                    Вход
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        </section>
        <section className="w-full py-12 md:py-24 lg:py-32 bg-muted">
          <div className="container px-4 md:px-6">
            <div className="mx-auto grid max-w-5xl items-center gap-6 lg:grid-cols-2 lg:gap-12">
              <div className="space-y-4">
                <div className="inline-block rounded-lg bg-gray-100 px-3 py-1 text-sm">Для пользователей</div>
                <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl">Проходите опросы</h2>
                <p className="text-gray-500 md:text-xl/relaxed lg:text-base/relaxed xl:text-xl/relaxed">
                  Участвуйте в опросах, делитесь своим мнением и получайте результаты.
                </p>
              </div>
              <div className="space-y-4">
                <div className="inline-block rounded-lg bg-gray-100 px-3 py-1 text-sm">Для создателей</div>
                <h2 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl">Создавайте опросы</h2>
                <p className="text-gray-500 md:text-xl/relaxed lg:text-base/relaxed xl:text-xl/relaxed">
                  Легко создавайте опросы, добавляйте вопросы и варианты ответов, анализируйте результаты.
                </p>
              </div>
            </div>
          </div>
        </section>
        <section className="w-full py-12 md:py-24 lg:py-32">
          <div className="container px-4 md:px-6">
            <div className="mx-auto grid max-w-5xl gap-6 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>Регистрация</CardTitle>
                  <CardDescription>Создайте аккаунт для доступа к опросам</CardDescription>
                </CardHeader>
                <CardContent>
                  <p>Быстрая регистрация с использованием имени пользователя и пароля.</p>
                </CardContent>
                <CardFooter>
                  <Link href="/register" className="w-full">
                    <Button className="w-full">Регистрация</Button>
                  </Link>
                </CardFooter>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Вход</CardTitle>
                  <CardDescription>Войдите в свой аккаунт</CardDescription>
                </CardHeader>
                <CardContent>
                  <p>Авторизуйтесь для доступа к опросам и созданию новых.</p>
                </CardContent>
                <CardFooter>
                  <Link href="/login" className="w-full">
                    <Button className="w-full" variant="outline">
                      Вход
                    </Button>
                  </Link>
                </CardFooter>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Опросы</CardTitle>
                  <CardDescription>Просмотр доступных опросов</CardDescription>
                </CardHeader>
                <CardContent>
                  <p>Просматривайте список доступных опросов и принимайте участие.</p>
                </CardContent>
                <CardFooter>
                  <Link href="/surveys" className="w-full">
                    <Button className="w-full" variant="secondary">
                      Опросы
                    </Button>
                  </Link>
                </CardFooter>
              </Card>
            </div>
          </div>
        </section>
      </main>
    </div>
  )
}

