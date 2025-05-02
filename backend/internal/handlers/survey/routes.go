package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterSurveyRoutes(r *gin.Engine, db *sql.DB) {
    r.GET("/surveys/:survey_id/questions", getQuestions(db))
    r.GET("/surveys", getSurveys(db))
    r.POST("/survey", createSurvey(db))
    r.DELETE("/survey/:id", deleteSurvey(db)) 
    r.POST("/question", createQuestion(db))
    r.DELETE("/question/:id", deleteQuestion(db)) 
    r.POST("/option", createAnswerUser(db))
    r.GET("/survey/:survey_id/answers/simple", getUserSimpleSurveyAnswers(db))
    r.PUT("/survey/:id", updateSurvey(db))
    r.GET("/survey/:survey_id/results", GetSurveyResultsHandler(db))
}
