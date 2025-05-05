"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { MainNav } from "@/components/main-nav"
import { useToast } from "@/components/ui/use-toast"
import Link from "next/link"
import { Loader2 } from "lucide-react"

const formSchema = z.object({
  username: z.string().min(1, {
    message: "Введите имя пользователя",
  }),
  password: z.string().min(1, {
    message: "Введите пароль",
  }),
})

export default function LoginPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [isLoading, setIsLoading] = useState(false)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  })
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"
  
  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsLoading(true)
    try {
      const response = await fetch(`${apiUrl}/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(values),
      })

      if (!response.ok) {
        throw new Error("HTTP error " + response.status)
      }

      const data = await response.json()

      // Сохраняем токен
      localStorage.setItem("token", data.token)

      toast({
        title: "Авторизация успешна",
        description: "Вы успешно вошли в систему.",
      })

      router.push("/surveys")
    } catch (error) {
      console.error("Ошибка авторизации:", error)
      toast({
        title: "Ошибка авторизации",
        description: "Неверное имя пользователя или пароль.",
        variant: "destructive",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-2xl">Вход</CardTitle>
            <CardDescription>Войдите в свой аккаунт для доступа к опросам</CardDescription>
          </CardHeader>
          <CardContent>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Имя пользователя</FormLabel>
                      <FormControl>
                        <Input placeholder="Введите имя пользователя" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Пароль</FormLabel>
                      <FormControl>
                        <Input type="password" placeholder="Введите пароль" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button type="submit" className="w-full" disabled={isLoading}>
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Вход...
                    </>
                  ) : (
                    "Войти"
                  )}
                </Button>
              </form>
            </Form>
          </CardContent>
          <CardFooter className="flex justify-center">
            <p className="text-sm text-muted-foreground">
              Нет аккаунта?{" "}
              <Link href="/register" className="text-primary hover:underline">
                Зарегистрироваться
              </Link>
            </p>
          </CardFooter>
        </Card>
      </main>
    </div>
  )
}

