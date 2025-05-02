package database

import (
    "database/sql"
    "errors"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "my-auth-app/internal/models"
)

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