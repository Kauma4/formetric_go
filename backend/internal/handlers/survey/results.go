package handlers

import (
    "errors"
    "database/sql"
    "strconv"
    "net/http"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
)


func GetSurveyResultsHandler(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        surveyID, err := strconv.Atoi(c.Param("survey_id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        userID, err := getUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        results, err := database.GetSurveyResults(db, surveyID, userID)
        if err != nil {
            if errors.Is(err, sql.ErrNoRows) {
                c.JSON(http.StatusNotFound, gin.H{"error": "Результаты не найдены"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, results)
    }
}