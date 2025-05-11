package database

import (
    "database/sql"
    "errors"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "my-auth-app/internal/models"
    "my-auth-app/internal/utils"
)
/*
func CreateUser(db *sql.DB, user *models.User) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return errors.New("failed to hash password")
    }

    _, err = db.Exec(
        "INSERT INTO users (username, password, email, full_name, avatar_url, phone_number, date_of_birth, location) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
        user.Username, string(hashedPassword), user.Email, user.FullName, user.AvatarURL, 
        user.PhoneNumber, user.DateOfBirth, user.Location,
    )
    return err
}
*/

func CreateUser(db *sql.DB, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	_, err = db.Exec(
		"INSERT INTO users (username, password, email, full_name, avatar_url, phone_number, date_of_birth, location) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		user.Username, string(hashedPassword), user.Email, user.FullName, user.AvatarURL,
		user.PhoneNumber, user.DateOfBirth, user.Location,
	)
	if err != nil {
		return err
	}

	// Отправляем письмо с подтверждением, если email указан
	if user.Email.Valid && user.Email.String != "" {
		err = utils.SendVerificationEmail(user.Email.String, user.Username)
		if err != nil {
			// Логируем ошибку, но не прерываем процесс, так как регистрация уже успешна
			fmt.Printf("Failed to send verification email to %s: %v\n", user.Email.String, err)
		}
	}

	return nil
}

func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
    var user models.User
    err := db.QueryRow(
        "SELECT id, username, password, email, full_name, avatar_url, phone_number, date_of_birth, location FROM users WHERE username = $1",
        username,
    ).Scan(
        &user.ID, &user.Username, &user.Password, &user.Email, &user.FullName,
        &user.AvatarURL, &user.PhoneNumber, &user.DateOfBirth, &user.Location,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, errors.New("database error")
    }
    return &user, nil
}

func GetUserByLogin(db *sql.DB, login string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, password, email, full_name, avatar_url, phone_number, date_of_birth, location 
		FROM users 
		WHERE username = $1 OR email = $1
	`
	err := db.QueryRow(query, login).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email, &user.FullName,
		&user.AvatarURL, &user.PhoneNumber, &user.DateOfBirth, &user.Location,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}


func GetUserByID(db *sql.DB, userID int) (*models.User, error) {
    var user models.User
    err := db.QueryRow(
        "SELECT id, username, email, full_name, avatar_url, phone_number, date_of_birth, location FROM users WHERE id = $1",
        userID,
    ).Scan(
        &user.ID, &user.Username, &user.Email, &user.FullName,
        &user.AvatarURL, &user.PhoneNumber, &user.DateOfBirth, &user.Location,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func UpdateUser(db *sql.DB, user *models.User) error {
    _, err := db.Exec(
        `UPDATE users SET username=$1, password=$2, email=$3, full_name=$4, 
        avatar_url=$5, phone_number=$6, date_of_birth=$7, location=$8 WHERE id=$9`,
        user.Username, user.Password, user.Email, user.FullName,
        user.AvatarURL, user.PhoneNumber, user.DateOfBirth, user.Location, user.ID,
    )
    return err
}

func UserExists(db *sql.DB, userID int) (bool, error) {
    var exists bool
    err := db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)", 
        userID,
    ).Scan(&exists)
    return exists, err
}

func GetUserAnswersRaw(db *sql.DB, userID, surveyID int) (*sql.Rows, error) {
    query := `
        SELECT 
            q.id,
            q.question_text, 
            q.answers, 
            q.correct_answer, 
            q.is_test,
            au.answer_id, 
            au.answer_user 
        FROM answers_users au
        JOIN questions q ON au.question_id = q.id
        WHERE au.user_id = $1 AND q.survey_id = $2`
    
    return db.Query(query, userID, surveyID)
}

func SaveUserResult(db *sql.DB, surveyID, userID, totalBall int) error {
    _, err := db.Exec(
        `INSERT INTO results (survey_id, user_id, total_ball) 
        VALUES ($1, $2, $3)
        ON CONFLICT (survey_id, user_id) 
        DO UPDATE SET total_ball = $3, date = NOW()`,
        surveyID, userID, totalBall,
    )
    return err
}

func DeleteUserAnswers(db *sql.DB, userID, surveyID int) error {
    _, err := db.Exec(
        `DELETE FROM answers_users 
        WHERE user_id = $1 AND question_id IN (
            SELECT id FROM questions WHERE survey_id = $2
        )`,
        userID, surveyID,
    )
    return err
}

func DeleteUserResult(db *sql.DB, userID, surveyID int) error {
    _, err := db.Exec(
        "DELETE FROM results WHERE user_id = $1 AND survey_id = $2",
        userID, surveyID,
    )
    return err
}

// Для результатов пользователя
func GetUserResults(db *sql.DB, userID int) ([]models.Result, error) {
    query := `
        SELECT id, survey_id, date, total_ball 
        FROM results 
        WHERE user_id = $1 
        ORDER BY date DESC`

    rows, err := db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
    }
    defer rows.Close()

    var results []models.Result
    for rows.Next() {
        var r models.Result
        if err := rows.Scan(&r.ID, &r.SurveyID, &r.Date, &r.TotalBall); err != nil {
            return nil, fmt.Errorf("ошибка сканирования результата: %v", err)
        }
        r.UserID = userID
        results = append(results, r)
    }

    if len(results) == 0 {
        return nil, sql.ErrNoRows
    }
    return results, nil
}

// UserExistsRaw проверяет существование пользователя без учета deleted_at
func UserExistsRaw(db *sql.DB, userID int) (bool, error) {
    var exists bool
    err := db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", 
        userID,
    ).Scan(&exists)
    return exists, err
}

func SoftDeleteUser(tx *sql.Tx, userID int) (int64, error) {
    result, err := tx.Exec(
        `UPDATE users SET deleted_at=NOW(), email=NULL, phone_number=NULL, 
        username=CONCAT(username, '_deleted_', EXTRACT(EPOCH FROM NOW())) 
        WHERE id=$1 AND deleted_at IS NULL`,
        userID,
    )
    if err != nil {
        return 0, fmt.Errorf("ошибка soft delete: %v", err)
    }
    return result.RowsAffected()
}