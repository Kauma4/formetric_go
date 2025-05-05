package models

import "database/sql"
import "time"

type User struct {
    ID          int       `json:"id"`
    Username    string    `json:"username"`
    Password    string    `json:"password"`
    Email       sql.NullString   `json:"email"`
    FullName    sql.NullString   `json:"full_name"`
    AvatarURL   sql.NullString   `json:"avatar_url"`
    PhoneNumber sql.NullString   `json:"phone_number"`
    DateOfBirth sql.NullString   `json:"date_of_birth"`
    Location    sql.NullString   `json:"location"`
}

type ParticipantResult struct {
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    TotalBall int       `json:"total_score"`
    Date      time.Time `json:"date"`
    MaxScore  int       `json:"max_score"`
}

type Survey struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    IsPrivate   bool   `json:"is_private"` 
    CreatedBy   int    `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
}

type SurveyResult struct {
    TotalScore int    `json:"total_score"`
    MaxScore   int    `json:"max_score"`
    Date       string `json:"date"`
}

type Question struct {
    ID           int      `json:"id"`
    SurveyID     int      `json:"survey_id"`
	IsRequired    bool     `json:"is_required"`
    IsTest        bool     `json:"is_test"`
    CorrectAnswer string   `json:"correct_answer,omitempty"`
    QuestionText string   `json:"question_text"`
    Ball         int      `json:"ball"`
    Answers      []Answer `json:"answers"`
}

type Answer struct {
    Text    string `json:"text"`
    Correct bool   `json:"correct"`
}

type AnswerUser struct {
	ID         int `json:"id"`
	UserID     int `json:"user_id"`
	QuestionID int `json:"question_id"`
	AnswerID   int `json:"answer_id"`
	AnswerText string `json:"answer_user"`
}

type Result struct {
	ID        int    `json:"id"`
	SurveyID  int    `json:"survey_id"`
	UserID    int    `json:"user_id"`
	Date      string `json:"date"`
	TotalBall int    `json:"total_ball"`
}

type ExtendedResult struct {
    Result
    Username string `json:"username"`
}

type AnswersRequest struct {
    Answers []AnswerUser `json:"answers"`
}

type UserAnswerSimple struct {
    QuestionText   string `json:"question_text"`
    UserAnswer     string `json:"user_answer"`
    CorrectAnswer  string `json:"correct_answer,omitempty"` // Добавляем для всех типов вопросов
    IsCorrect      *bool  `json:"is_correct,omitempty"`     // Только для тестовых вопросов
}





func (q *Question) GetQuestionType() string {
    if q.IsTest {
        return "test"
    }
    if q.CorrectAnswer != "" {
        return "text_with_answer"
    }
    return "text_info"
}