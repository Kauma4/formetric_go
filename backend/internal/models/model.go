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

// RegisterUserInput используется для обработки входных данных при регистрации
type RegisterUserInput struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email,omitempty"`
	FullName    string `json:"full_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	Location    string `json:"location,omitempty"`
}

// ToUser преобразует RegisterUserInput в User
func (input *RegisterUserInput) ToUser() User {
	return User{
		Username:    input.Username,
		Password:    input.Password,
		Email:       sql.NullString{String: input.Email, Valid: input.Email != ""},
		FullName:    sql.NullString{String: input.FullName, Valid: input.FullName != ""},
		AvatarURL:   sql.NullString{String: input.AvatarURL, Valid: input.AvatarURL != ""},
		PhoneNumber: sql.NullString{String: input.PhoneNumber, Valid: input.PhoneNumber != ""},
		DateOfBirth: sql.NullString{String: input.DateOfBirth, Valid: input.DateOfBirth != ""},
		Location:    sql.NullString{String: input.Location, Valid: input.Location != ""},
	}
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
    MaxBall     int       `json:"max_ball"`
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
    MultipleAnswers bool `json:"multipleAnswers"`
    IsTest        bool     `json:"is_test"`
    CorrectAnswer string   `json:"correct_answer,omitempty"`
    QuestionText string   `json:"question_text"`
    Ball         int      `json:"ball"`
    Answers      []Answer `json:"answers"`
}
/*
type Answer struct {
    Text    string `json:"text"`
    Correct bool   `json:"correct"`
}
*/
type Answer struct {
    ID      int    `json:"id,omitempty"`
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
    CorrectAnswer  string `json:"correct_answer,omitempty"`
    IsCorrect      *bool  `json:"is_correct,omitempty"`
}


// SurveyAnalytics содержит полную аналитику по опросу
type SurveyAnalytics struct {
	Survey        Survey                    `json:"survey"`
	Participants  []ParticipantResult       `json:"participants"`
	Questions     []QuestionAnalytics       `json:"questions"`
	TotalResponses int                      `json:"total_responses"`
	AverageScore   float64                  `json:"average_score"`
}

// QuestionAnalytics содержит статистику по конкретному вопросу
type QuestionAnalytics struct {
	Question       Question                 `json:"question"`
	AnswerStats    []AnswerStat             `json:"answer_stats"` // Для тестовых вопросов
	TextResponses  []TextResponse           `json:"text_responses"` // Для текстовых вопросов
	CorrectRate    float64                  `json:"correct_rate"` // % правильных ответов (для тестовых)
}

// AnswerStat статистика по вариантам ответов тестового вопроса
type AnswerStat struct {
	AnswerText string `json:"answer_text"`
	Count      int    `json:"count"`
	IsCorrect  bool   `json:"is_correct"`
}

// TextResponse текстовый ответ пользователя
type TextResponse struct {
	Username string `json:"username"`
	Answer   string `json:"answer"`
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