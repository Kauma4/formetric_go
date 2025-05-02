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

func GetSurveyResults(db *sql.DB, surveyID int, ownerID int) ([]models.ExtendedResult, error) {
    query := `
        SELECT r.id, r.user_id, r.date, r.total_ball, u.username
        FROM results r
        JOIN users u ON r.user_id = u.id
        WHERE r.survey_id = $1 AND EXISTS (
            SELECT 1 FROM surveys WHERE id = $1 AND created_by = $2
        )
        ORDER BY r.date DESC`

    rows, err := db.Query(query, surveyID, ownerID)
    if err != nil {
        return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
    }
    defer rows.Close()

    var results []models.ExtendedResult
    for rows.Next() {
        var er models.ExtendedResult
        if err := rows.Scan(&er.ID, &er.UserID, &er.Date, &er.TotalBall, &er.Username); err != nil {
            return nil, fmt.Errorf("ошибка сканирования результата: %v", err)
        }
        er.SurveyID = surveyID
        results = append(results, er)
    }

    if len(results) == 0 {
        return nil, sql.ErrNoRows
    }
    return results, nil
}