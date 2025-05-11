package handlers

import (
    "database/sql"
    "fmt"
	"log"
    "strings"
    "my-auth-app/internal/models"
	"my-auth-app/internal/database"
)

type Executor interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Query(query string, args ...interface{}) (*sql.Rows, error)
}

func checkAnswerCorrectness(answer models.AnswerUser, question *models.Question) (bool, error) {
    if question.IsTest {
        if answer.AnswerID <= 0 || answer.AnswerID > len(question.Answers) {
            return false, fmt.Errorf("неверный ID ответа для вопроса %d", question.ID)
        }
        return question.Answers[answer.AnswerID-1].Correct, nil
    }
    
    if question.CorrectAnswer != "" {
        return strings.EqualFold(
            strings.TrimSpace(answer.AnswerText),
            strings.TrimSpace(question.CorrectAnswer)), nil
    }
    
    return false, nil
}

func processCorrectAnswer(db *sql.DB, question *models.Question, answer models.AnswerUser) error {
    surveyID, err := database.GetSurveyIDByQuestionID(db, answer.QuestionID)
    if err != nil {
        log.Printf("Failed to get survey ID: %v", err)
        return fmt.Errorf("failed to get survey ID: %v", err)
    }
    
    return database.UpdateUserResult(db, surveyID, answer.UserID, question.Ball)
}
