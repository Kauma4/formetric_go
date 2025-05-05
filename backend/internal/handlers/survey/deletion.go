package handlers

import (
    "database/sql"
    "net/http"
    "strconv"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
)

func deleteSurvey(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        surveyIDStr := c.Param("id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        userID, err := getUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        if err := database.DeleteSurvey(db, surveyID, userID); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Опрос успешно удален",
            "deleted_id": surveyID,
        })
    }
}

func deleteQuestion(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        questionIDStr := c.Param("id")
        questionID, err := strconv.Atoi(questionIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID вопроса"})
            return
        }

        userID, err := getUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        if err := database.DeleteQuestion(db, questionID, userID); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Вопрос успешно удален",
            "deleted_id": questionID,
        })
    }
}