package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

// RegisterSurveyRoutes регистрирует маршруты для опросов
func RegisterSurveyRoutes(r *gin.Engine, db *sql.DB) {
	// Группа маршрутов с префиксом /surveys
	surveysGroup := r.Group("/surveys")
	{
		surveysGroup.GET("/:id/questions", getQuestions(db))
		surveysGroup.GET("", getSurveys(db))
		surveysGroup.GET("/:id/participants", getSurveyParticipants(db))
	}

	// Группа маршрутов с префиксом /survey
	surveyGroup := r.Group("/survey")
	{
		surveyGroup.POST("", createSurvey(db))
		surveyGroup.DELETE("/:id", deleteSurvey(db))
		surveyGroup.PUT("/:id", updateSurvey(db))
		surveyGroup.GET("/:id/answers/simple", getUserSimpleSurveyAnswers(db))
		surveyGroup.POST("/:id/answers/user", getUserAnswersByID(db))
		surveyGroup.GET("/:id/results", GetSurveyResultsHandler(db))
	}

	// Маршруты для вопросов и ответов
	r.POST("/question", createQuestion(db))
	r.DELETE("/question/:id", deleteQuestion(db))
	r.POST("/option", createAnswerUser(db))

	// Маршрут для пользовательских опросов
	r.GET("/my-surveys", getUserSurveys(db))
}