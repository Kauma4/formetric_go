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
  username: z.string().min(3, {
    message: "Имя пользователя должно содержать не менее 3 символов",
  }),
  password: z.string().min(6, {
    message: "Пароль должен содержать не менее 6 символов",
  }),
})

export default function RegisterPage() {
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

 // const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/"

 const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsLoading(true)
    try {
      const response = await fetch(`${apiUrl}/register`, {  
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
      toast({
        title: "Регистрация успешна",
        description: "Вы успешно зарегистрировались. Теперь вы можете войти в систему.",
      })
      router.push("/login")
    } catch (error) {
      console.error("Ошибка регистрации:", error)
      toast({
        title: "Ошибка регистрации",
        description: "Произошла ошибка при регистрации. Пожалуйста, попробуйте еще раз.",
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
            <CardTitle className="text-2xl">Регистрация</CardTitle>
            <CardDescription>Создайте аккаунт для доступа к опросам</CardDescription>
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
                      Регистрация...
                    </>
                  ) : (
                    "Зарегистрироваться"
                  )}
                </Button>
              </form>
            </Form>
          </CardContent>
          <CardFooter className="flex justify-center">
            <p className="text-sm text-muted-foreground">
              Уже есть аккаунт?{" "}
              <Link href="/login" className="text-primary hover:underline">
                Войти
              </Link>
            </p>
          </CardFooter>
        </Card>
      </main>
    </div>
  )
}

