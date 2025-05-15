package jwt

import (
    "errors"
    "github.com/dgrijalva/jwt-go"
    "time"
    "log"
    "os"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID int, username string) (string, error) {
    if len(jwtSecret) == 0 {
        log.Println("JWT_SECRET не задано в переменных окружения")
        return "", errors.New("JWT_SECRET не задано")
    }

    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      time.Now().Add(time.Hour * 24).Unix(), // Токен действует 24 часа
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        log.Println("Ошибка при генерации JWT токена:", err)
        return "", err
    }

    return tokenString, nil
}

func GenerateJWTWithExpiration(userID int, username string, expiresAt time.Time) (string, error) {
    if len(jwtSecret) == 0 {
        return "", errors.New("JWT_SECRET не задано")
    }
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "exp":      expiresAt.Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ExtractUserIDFromJWT(tokenString string) (int, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return 0, errors.New("ошибка при разборе токена: " + err.Error())
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if userID, ok := claims["user_id"].(float64); ok {
            return int(userID), nil
        }
        return 0, errors.New("не найден user_id в токене")
    }

    return 0, errors.New("невалидный токен")
}