"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { useEffect, useState } from "react"
import { LogOut, Menu } from 'lucide-react'
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"

export function MainNav() {
  const pathname = usePathname()
  const router = useRouter()
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    const token = localStorage.getItem("token")
    setIsLoggedIn(!!token)
  }, [])

  const handleLogout = () => {
    localStorage.removeItem("token")
    setIsLoggedIn(false)
    router.push("/login")
  }

  const routes = [
    { href: "/", label: "Главная", showWhenLoggedOut: true },
    { href: "/surveys", label: "Опросы", showAlways: true },
    { href: "/my-surveys", label: "Мои опросы", showWhenLoggedIn: true }, // Добавлено
    { href: "/create-survey", label: "Создать опрос", showWhenLoggedIn: true },
    { href: "/analytics", label: "Аналитика", showWhenLoggedIn: true },
    { href: "/account", label: "Личный кабинет", showWhenLoggedIn: true },
    { href: "/login", label: "Вход", showWhenLoggedOut: true },
    { href: "/register", label: "Регистрация", showWhenLoggedOut: true },
  ]

  const filteredRoutes = routes.filter((route) => {
    if (route.showAlways) return true
    if (route.showWhenLoggedIn && isLoggedIn) return true
    if (route.showWhenLoggedOut && !isLoggedIn) return true
    return false
  })

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center">
        <div className="mr-4 hidden md:flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <span className="font-bold">Опросы</span>
          </Link>
          <nav className="flex items-center space-x-6 text-sm font-medium">
            {filteredRoutes.map((route) => (
              <Link
                key={route.href}
                href={route.href}
                className={`transition-colors hover:text-foreground/80 ${
                  pathname === route.href ? "text-foreground" : "text-foreground/60"
                }`}
              >
                {route.label}
              </Link>
            ))}
          </nav>
        </div>
        <Sheet open={isOpen} onOpenChange={setIsOpen}>
          <SheetTrigger asChild className="md:hidden">
            <Button variant="ghost" size="icon" className="mr-2">
              <Menu className="h-5 w-5" />
              <span className="sr-only">Toggle Menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="pr-0">
            <Link href="/" className="flex items-center" onClick={() => setIsOpen(false)}>
              <span className="font-bold">Опросы</span>
            </Link>
            <nav className="mt-6 flex flex-col space-y-4">
              {filteredRoutes.map((route) => (
                <Link
                  key={route.href}
                  href={route.href}
                  onClick={() => setIsOpen(false)}
                  className={`text-sm font-medium transition-colors hover:text-foreground/80 ${
                    pathname === route.href ? "text-foreground" : "text-foreground/60"
                  }`}
                >
                  {route.label}
                </Link>
              ))}
              {isLoggedIn && (
                <Button variant="ghost" className="justify-start px-2" onClick={handleLogout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  Выйти
                </Button>
              )}
            </nav>
          </SheetContent>
        </Sheet>
        <div className="flex flex-1 items-center justify-end space-x-4">
          <nav className="flex items-center space-x-2">
            {isLoggedIn ? (
              <Button variant="ghost" size="sm" onClick={handleLogout} className="hidden md:flex">
                <LogOut className="mr-2 h-4 w-4" />
                Выйти
              </Button>
            ) : (
              <div className="hidden space-x-2 md:flex">
                <Link href="/login">
                  <Button variant="ghost" size="sm">
                    Вход
                  </Button>
                </Link>
                <Link href="/register">
                  <Button size="sm">Регистрация</Button>
                </Link>
              </div>
            )}
          </nav>
        </div>
      </div>
    </header>
  )
}

