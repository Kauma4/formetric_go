package main

import (
	"log"
	"my-auth-app/internal/database"

	authHandlers "my-auth-app/internal/handlers/auth"
	surveyHandlers "my-auth-app/internal/handlers/survey"
	
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {

	// Используем один экземпляр gin.Default()
	router := gin.Default()
	// Включаем CORS
	router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"}, // Разрешаем только с этого адреса
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    }))

	// Подключаем базу данных
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	// Обслуживание статических файлов из директории uploads
	router.Static("/uploads", "./uploads")

	// Регистрируем маршруты
	authHandlers.RegisterRoutes(router, db)
	surveyHandlers.RegisterSurveyRoutes(router, db)

	// Запускаем сервер
	router.Run(":8080")
}

