
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
)
func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	router := gin.Default()
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

	router.Static("/uploads", "./uploads")


	analyticsHandlers.RegisterAnalyticsRoutes(router, db)
	authHandlers.RegisterRoutes(router, db)
	surveyHandlers.RegisterSurveyRoutes(router, db)

	router.Run(":8080")
}
