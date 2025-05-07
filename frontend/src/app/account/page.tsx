"use client"

import { useEffect, useState, useRef } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { MainNav } from "@/components/main-nav"
import { useToast } from "@/components/ui/use-toast"
import { Loader2, Pencil, LogOut, Camera, Eye, EyeOff } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface UserData {
  id: number
  username: string
  email?: string
  avatar_url?: string
  full_name?: string
  phone_number?: string
  date_of_birth?: string
  location?: string
}

const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
const phoneRegex = /^\+7[-\s]?\(?\d{3}\)?[-\s]?\d{3}[-\s]?\d{2}[-\s]?\d{2}$/;
const maxFileSize = 2 * 1024 * 1024; // 2 MB
const allowedFileTypes = ["image/jpeg", "image/png", "image/gif"];
const BASE_URL = "http://localhost:8080"; // Базовый URL сервера

export default function AccountPage() {
  const [userData, setUserData] = useState<UserData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    full_name: "",
    phone_number: "",
    date_of_birth: "",
    location: ""
  })
  const [avatarPreview, setAvatarPreview] = useState<string>("")
  const [isAvatarLoading, setIsAvatarLoading] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const { toast } = useToast()
  const router = useRouter()
  const avatarInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const fetchUserData = async () => {
      const token = localStorage.getItem("token")
      if (!token) {
        router.push("/login")
        return
      }

      try {
        const response = await fetch(`${BASE_URL}/user`, {
          headers: { Authorization: `Bearer ${token}` }
        })
        
        if (!response.ok) throw new Error("Ошибка загрузки данных пользователя")

        const data = await response.json()
        console.log("Fetched user data:", data) // Логирование данных пользователя
        setUserData(data)
        setFormData({
          username: data.username,
          email: data.email || "",
          password: "",
          full_name: data.full_name || "",
          phone_number: data.phone_number || "",
          date_of_birth: data.date_of_birth?.split('T')[0] || "",
          location: data.location || ""
        })
      } catch (error) {
        toast({ variant: "destructive", title: "Ошибка", description: "Не удалось загрузить данные пользователя" })
      } finally {
        setIsLoading(false)
      }
    }

    fetchUserData()

    return () => {
      // Очистка временного URL предварительного просмотра
      if (avatarPreview) URL.revokeObjectURL(avatarPreview)
    }
  }, [router, toast])

  const handleUpdate = async () => {
    const token = localStorage.getItem("token")
    if (!token) {
      toast({ variant: "destructive", title: "Ошибка", description: "Требуется авторизация" })
      return
    }

    if (!formData.username) {
      toast({ variant: "destructive", title: "Ошибка", description: "Имя пользователя обязательно" })
      return
    }

    if (formData.email && !emailRegex.test(formData.email)) {
      toast({ variant: "destructive", title: "Ошибка", description: "Неверный формат email" })
      return
    }

    if (formData.phone_number && !phoneRegex.test(formData.phone_number)) {
      toast({ variant: "destructive", title: "Ошибка", description: "Неверный формат телефона" })
      return
    }

    try {
      const response = await fetch(`${BASE_URL}/user`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify(formData)
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || "Ошибка обновления профиля")
      }

      const updatedData = await response.json()
      setUserData(updatedData)
      setIsEditing(false)
      toast({ title: "Успех", description: "Данные профиля обновлены" })
    } catch (error: any) {
      toast({ variant: "destructive", title: "Ошибка", description: error.message || "Не удалось обновить данные" })
    }
  }

  const handleAvatarChange = async (files: FileList | null) => {
    if (!files || files.length === 0) return

    const file = files[0]
    if (!allowedFileTypes.includes(file.type)) {
      toast({ variant: "destructive", title: "Ошибка", description: "Разрешены только файлы JPEG, PNG или GIF" })
      return
    }

    if (file.size > maxFileSize) {
      toast({ variant: "destructive", title: "Ошибка", description: "Размер файла не должен превышать 2 МБ" })
      return
    }

    const token = localStorage.getItem("token")
    if (!token) {
      toast({ variant: "destructive", title: "Ошибка", description: "Требуется авторизация" })
      return
    }

    const formData = new FormData()
    formData.append("avatar", file)

    setIsAvatarLoading(true)
    try {
      const response = await fetch(`${BASE_URL}/user/avatar`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`
        },
        body: formData
      })

      const data = await response.json()
      console.log("Avatar upload response:", data) // Логирование ответа
      if (!response.ok) {
        throw new Error(data.error || "Ошибка загрузки аватара")
      }

      // Повторный запрос данных пользователя для синхронизации
      const userResponse = await fetch(`${BASE_URL}/user`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!userResponse.ok) throw new Error("Ошибка загрузки данных пользователя")
      const updatedUserData = await userResponse.json()
      console.log("Updated user data after avatar upload:", updatedUserData) // Логирование обновленных данных
      setUserData(updatedUserData)

      setAvatarPreview("")
      toast({ title: "Успех", description: "Аватар успешно обновлен" })
    } catch (error: any) {
      console.error("Avatar upload error:", error)
      toast({ variant: "destructive", title: "Ошибка", description: error.message || "Не удалось загрузить аватар" })
    } finally {
      setIsAvatarLoading(false)
      if (avatarInputRef.current) avatarInputRef.current.value = ""
    }
  }

  const handleLogout = () => {
    localStorage.removeItem("token")
    router.push("/login")
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  if (!userData) {
    return (
      <div className="flex justify-center items-center h-screen">
        <p>Не удалось загрузить данные пользователя</p>
      </div>
    )
  }

  const hasChanges = JSON.stringify({
    username: userData.username,
    email: userData.email || "",
    full_name: userData.full_name || "",
    phone_number: userData.phone_number || "",
    date_of_birth: userData.date_of_birth?.split('T')[0] || "",
    location: userData.location || ""
  }) !== JSON.stringify({
    username: formData.username,
    email: formData.email,
    full_name: formData.full_name,
    phone_number: formData.phone_number,
    date_of_birth: formData.date_of_birth,
    location: formData.location
  });

  return (
    <div className="flex min-h-screen flex-col">
      <MainNav />
      <main className="flex-1 container py-6">
        <Card className="max-w-2xl mx-auto">
          <CardHeader className="flex flex-row justify-between items-center">
            <CardTitle className="text-2xl">Личный кабинет</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="flex flex-col items-center gap-4">
              <div 
                className="relative group w-32 h-32 rounded-full cursor-pointer"
                onClick={() => avatarInputRef.current?.click()}
              >
                {isAvatarLoading ? (
                  <div className="w-full h-full rounded-full bg-gray-100 flex items-center justify-center">
                    <Loader2 className="h-8 w-8 animate-spin" />
                  </div>
                ) : (
                  <>
                    <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity rounded-full flex items-center justify-center">
                      <Camera className="text-white w-8 h-8" />
                    </div>
                    {userData.avatar_url || avatarPreview ? (
                      <img
                        src={avatarPreview || `${BASE_URL}${userData.avatar_url}`}
                        alt="Аватар"
                        className="w-full h-full rounded-full object-cover border-2"
                        onError={(e) => console.error("Error loading avatar:", userData.avatar_url)} // Логирование ошибок загрузки
                      />
                    ) : (
                      <div className="w-full h-full rounded-full bg-gray-100 flex items-center justify-center text-4xl font-bold">
                        {userData.username[0].toUpperCase()}
                      </div>
                    )}
                  </>
                )}
              </div>

              <input
                id="avatarInput"
                ref={avatarInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(e) => handleAvatarChange(e.target.files)}
              />

              <span className="text-sm text-muted-foreground">
                Нажмите на аватар для изменения
              </span>
            </div>

            {isEditing ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>Имя пользователя</Label>
                  <Input
                    value={formData.username}
                    onChange={(e) => setFormData({...formData, username: e.target.value})}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Полное имя</Label>
                  <Input
                    value={formData.full_name}
                    onChange={(e) => setFormData({...formData, full_name: e.target.value})}
                    placeholder="Введите ваше полное имя"
                  />
                </div>
                
                <div className="space-y-2">
                  <Label>Email</Label>
                  <Input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({...formData, email: e.target.value})}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Телефон</Label>
                  <Input
                    type="tel"
                    value={formData.phone_number}
                    onChange={(e) => setFormData({...formData, phone_number: e.target.value})}
                    placeholder="+7 (XXX) XXX-XX-XX"
                  />
                </div>

                <div className="space-y-2">
                  <Label>Дата рождения</Label>
                  <Input
                    type="date"
                    value={formData.date_of_birth}
                    onChange={(e) => setFormData({...formData, date_of_birth: e.target.value})}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Местоположение</Label>
                  <Input
                    value={formData.location}
                    onChange={(e) => setFormData({...formData, location: e.target.value})}
                    placeholder="Город, страна"
                  />
                </div>

                <div className="space-y-2">
                  <Label>Новый пароль</Label>
                  <div className="flex items-center">
                    <Input
                      type={showPassword ? "text" : "password"}
                      value={formData.password}
                      onChange={(e) => setFormData({...formData, password: e.target.value})}
                      placeholder="Введите новый пароль"
                    />
                    <Button 
                      variant="outline"
                      className="ml-2"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? <EyeOff /> : <Eye />}
                    </Button>
                  </div>
                </div>

                <div className="flex gap-2">
                  <Button onClick={handleUpdate} disabled={!hasChanges}>Сохранить</Button>
                  <Button variant="outline" onClick={() => {
                    setIsEditing(false)
                    setFormData({
                      username: userData.username,
                      email: userData.email || "",
                      password: "",
                      full_name: userData.full_name || "",
                      phone_number: userData.phone_number || "",
                      date_of_birth: userData.date_of_birth?.split('T')[0] || "",
                      location: userData.location || ""
                    })
                  }}>
                    Отмена
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                {userData.full_name && (
                  <div className="space-y-2">
                    <h3 className="font-medium">Полное имя:</h3>
                    <p>{userData.full_name}</p>
                  </div>
                )}
                
                <div className="space-y-2">
                  <h3 className="font-medium">Имя пользователя:</h3>
                  <p>{userData.username}</p>
                </div>
                
                <div className="space-y-2">
                  <h3 className="font-medium">Email:</h3>
                  <p>{userData.email || "Не указан"}</p>
                </div>

                {userData.phone_number && (
                  <div className="space-y-2">
                    <h3 className="font-medium">Телефон:</h3>
                    <p>{userData.phone_number}</p>
                  </div>
                )}

                {userData.date_of_birth && (
                  <div className="space-y-2">
                    <h3 className="font-medium">Дата рождения:</h3>
                    <p>{new Date(userData.date_of_birth).toLocaleDateString()}</p>
                  </div>
                )}

                {userData.location && (
                  <div className="space-y-2">
                    <h3 className="font-medium">Местоположение:</h3>
                    <p>{userData.location}</p>
                  </div>
                )}

                <Button 
                  onClick={() => setIsEditing(true)}
                  className="gap-2"
                >
                  <Pencil className="h-4 w-4" />
                  Редактировать профиль
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  )
}