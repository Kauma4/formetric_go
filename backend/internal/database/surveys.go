package database

import (
    "fmt"
    "database/sql"
    "errors"
    "my-auth-app/internal/models"
)
func CreateSurvey(db *sql.DB, survey *models.Survey) error {
    err := db.QueryRow(
      `INSERT INTO surveys 
      (title, description, is_private, created_by) 
      VALUES ($1, $2, $3, $4) 
      RETURNING id`,
      survey.Title, 
      survey.Description,
      survey.IsPrivate,
      survey.CreatedBy,
    ).Scan(&survey.ID)
    return err
  }

func GetSurveys(db *sql.DB) ([]models.Survey, error) {
    rows, err := db.Query("SELECT id, title, description, created_by FROM surveys")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var surveys []models.Survey
    for rows.Next() {
        var s models.Survey
        if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.CreatedBy); err != nil {
            return nil, err
        }
        surveys = append(surveys, s)
    }
    return surveys, nil
}

func GetSurveyOwnerID(db *sql.DB, surveyID int) (int, error) {
    var ownerID int
    err := db.QueryRow("SELECT created_by FROM surveys WHERE id = $1", surveyID).Scan(&ownerID)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, errors.New("опрос не найден")
        }
        return 0, err
    }
    return ownerID, nil
}

func DeleteSurvey(db *sql.DB, surveyID, userID int) error {
    ownerID, err := GetSurveyOwnerID(db, surveyID)
    if err != nil {
        return err
    }
    if ownerID != userID {
        return errors.New("недостаточно прав для удаления опроса")
    }
    _, err = db.Exec("DELETE FROM surveys WHERE id = $1", surveyID)
    return err
}

func UpdateSurvey(db *sql.DB, surveyID int, title, description string) error {
    _, err := db.Exec(
        "UPDATE surveys SET title = $1, description = $2 WHERE id = $3",
        title, description, surveyID,
    )
    return err
}

func GetSurveysByCreator(db *sql.DB, userID int) ([]models.Survey, error) {
    query := `
        SELECT id, title, description, is_private, created_by, created_at 
        FROM surveys 
        WHERE created_by = $1
        ORDER BY created_at DESC
    `
    
    rows, err := db.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var surveys []models.Survey
    for rows.Next() {
        var s models.Survey
        err := rows.Scan(
            &s.ID,
            &s.Title,
            &s.Description,
            &s.IsPrivate,
            &s.CreatedBy,
            &s.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        surveys = append(surveys, s)
    }
    
    return surveys, nil
}


func GetSurveyResults(db *sql.DB, surveyID int, userID int) (*models.SurveyResult, error) {
    var result models.SurveyResult
    
    // Запрос с объединением таблиц results и surveys
    err := db.QueryRow(`
        SELECT r.total_ball, s.max_ball, r.date 
        FROM results r
        JOIN surveys s ON r.survey_id = s.id
        WHERE r.survey_id = $1 AND r.user_id = $2
    `, surveyID, userID).Scan(
        &result.TotalScore,
        &result.MaxScore,
        &result.Date,
    )

    if err != nil {
        return nil, err
    }
    
    return &result, nil
}

func UpdateSurveyMaxBall(db Executor, surveyID int, ball int) error {
	_, err := db.Exec(
		"UPDATE surveys SET max_ball = max_ball + $1 WHERE id = $2",
		ball,
		surveyID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления max_ball: %v", err)
	}
	return nil
}

func GetSurveyParticipants(db *sql.DB, surveyID int) ([]models.ParticipantResult, error) {
    query := `
        SELECT 
            u.id,
            u.username,
            r.total_ball,
            r.date,
            s.max_ball
        FROM results r
        JOIN users u ON r.user_id = u.id
        JOIN surveys s ON r.survey_id = s.id
        WHERE r.survey_id = $1
        ORDER BY r.date DESC
    `
    
    rows, err := db.Query(query, surveyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var participants []models.ParticipantResult
    for rows.Next() {
        var p models.ParticipantResult
        err := rows.Scan(
            &p.UserID,
            &p.Username,
            &p.TotalBall,
            &p.Date,
            &p.MaxScore,
        )
        if err != nil {
            return nil, err
        }
        participants = append(participants, p)
    }
    
    return participants, nil
}




type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}