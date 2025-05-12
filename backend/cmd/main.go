package main

import (
	"log"
	"my-auth-app/internal/database"

	authHandlers "my-auth-app/internal/handlers/auth"
	surveyHandlers "my-auth-app/internal/handlers/survey"
	analyticsHandlers "my-auth-app/internal/handlers/analytics"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	router := gin.Default()
	
	// настрйка CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	// Статич файлы
	router.Static("/uploads", "./uploads")
	router.Static("/docs", "./docs")

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/docs/openapi.json"),
	))

	// Регистрация маршрутов
	analyticsHandlers.RegisterAnalyticsRoutes(router, db)
	authHandlers.RegisterRoutes(router, db)
	surveyHandlers.RegisterSurveyRoutes(router, db)

	// Запуск сервера
	router.Run(":8080")
}
