package handlers

import (
    "errors"
    "strings"
    "strconv"
    "log"
    "os"
    "io"
    "fmt"
    "time"
    "database/sql"
    "path/filepath" 
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
    "my-auth-app/internal/jwt"  
    "my-auth-app/internal/utils"
    "golang.org/x/crypto/bcrypt" 
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {
    r.Use(cors.Default())
    r.POST("/register", registerUser(db))
	r.POST("/login", loginUser(db))   
    r.GET("/user/results", GetUserResultsHandler(db))

    authGroup := r.Group("/")
    authGroup.Use(AuthMiddleware())
    {
        authGroup.GET("/user", getUser(db))
        authGroup.PUT("/user", updateUserProfile(db))
        authGroup.DELETE("/user", DeleteUserHandler(db))
        authGroup.POST("/user/avatar", uploadAvatar(db))
    }
}

func getUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем userID из токена
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        user, err := database.GetUserByID(db, userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
            return
        }
        if user == nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
            return
        }

        response := gin.H{
            "id":       user.ID,
            "username": user.Username,
        }
        
        if user.Email.Valid {
            response["email"] = user.Email.String
        }
        if user.FullName.Valid {
            response["full_name"] = user.FullName.String
        }
        if user.AvatarURL.Valid{
            response["avatar_url"] = user.AvatarURL.String
        }
        if user.PhoneNumber.Valid{
            response["phone_number"] = user.PhoneNumber.String
        }
        if user.DateOfBirth.Valid{
            response["date_of_birth"] = user.DateOfBirth.String
        }
        if user.Location.Valid{
            response["location"] = user.Location.String
        }

        c.JSON(http.StatusOK, response)
    }
}

func registerUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.RegisterUserInput

		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("Ошибка привязки JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос", "details": err.Error()})
			return
		}

		if input.Username == "" || input.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать имя пользователя и пароль"})
			return
		}

		existingUser, err := database.GetUserByUsername(db, input.Username)
		if err != nil {
			log.Printf("Ошибка проверки username: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}
		if existingUser != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Имя пользователя уже занято"})
			return
		}

		if input.Email != "" {
			existingUser, err := database.GetUserByLogin(db, input.Email)
			if err != nil {
				log.Printf("Ошибка проверки email: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
				return
			}
			if existingUser != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email уже используется"})
				return
			}
		}

		user := input.ToUser()
		if err := database.CreateUser(db, &user); err != nil {
			log.Printf("Ошибка создания пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Пользователь успешно зарегистрирован"})
	}
}

func loginUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials struct {
			Login    string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&credentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
			return
		}

		if credentials.Login == "" || credentials.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать логин и пароль"})
			return
		}

		user, err := database.GetUserByLogin(db, credentials.Login)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
			return
		}

		token, err := jwt.GenerateJWT(user.ID, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать токен"})
			return
		}

		// Если вход выполнен по email, отправляем токен на email
		isEmailLogin := user.Email.Valid && user.Email.String == credentials.Login
		if isEmailLogin {
			err = utils.SendTokenEmail(user.Email.String, user.Username, token)
			if err != nil {
				// Логируем ошибку, но не прерываем процесс, так как токен уже сгенерирован
				log.Printf("Failed to send token email to %s: %v\n", user.Email.String, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Успешный вход",
			"token":   token,
			"email_sent": isEmailLogin,
		})
	}
}

func uploadAvatar(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        file, header, err := c.Request.FormFile("avatar")
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось получить файл"})
            return
        }
        defer file.Close()

        allowedTypes := map[string]bool{
            "image/jpeg": true,
            "image/png":  true,
            "image/gif":  true,
        }
        if !allowedTypes[header.Header.Get("Content-Type")] {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Разрешены только файлы JPEG, PNG или GIF"})
            return
        }

        if header.Size > 2*1024*1024 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Размер файла не должен превышать 2 МБ"})
            return
        }

        uploadDir := "./uploads/avatars"
        if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
            log.Printf("Ошибка создания директории: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
            return
        }

        ext := filepath.Ext(header.Filename)
        filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
        filePath := filepath.Join(uploadDir, filename)

        out, err := os.Create(filePath)
        if err != nil {
            log.Printf("Ошибка создания файла: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
            return
        }
        defer out.Close()

        if _, err := io.Copy(out, file); err != nil {
            log.Printf("Ошибка копирования файла: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
            return
        }

        avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
        _, err = db.Exec(
            "UPDATE users SET avatar_url = $1 WHERE id = $2",
            avatarURL, userID,
        )
        if err != nil {
            log.Printf("Ошибка обновления avatar_url: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных пользователя"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Аватар успешно загружен",
            "avatar_url": avatarURL,
        })
    }
}

func updateUserProfile(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем userID из токена
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // Парсинг входных данных
        var updateData struct {
            Email       string `json:"email"`
            FullName    string `json:"full_name"`
            AvatarURL   string `json:"avatar_url"`
            PhoneNumber string `json:"phone_number"`
            DateOfBirth string `json:"date_of_birth"`
            Location    string `json:"location"`
        }

        if err := c.ShouldBindJSON(&updateData); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
            return
        }

        //  Подготовка данных для обновления
        updateFields := make(map[string]interface{})
        
        if updateData.Email != "" {
            updateFields["email"] = sql.NullString{String: updateData.Email, Valid: true}
        }
        
        if updateData.FullName != "" {
            updateFields["full_name"] = sql.NullString{String: updateData.FullName, Valid: true}
        }
        
        if updateData.AvatarURL != "" {
            updateFields["avatar_url"] = sql.NullString{String: updateData.AvatarURL, Valid: true}
        }
        
        if updateData.PhoneNumber != "" {
            updateFields["phone_number"] = sql.NullString{String: updateData.PhoneNumber, Valid: true}
        }
        
        if updateData.DateOfBirth != "" {
            if _, err := time.Parse("2006-01-02", updateData.DateOfBirth); err == nil {
                updateFields["date_of_birth"] = sql.NullString{String: updateData.DateOfBirth, Valid: true}
            }
        }
        
        if updateData.Location != "" {
            updateFields["location"] = sql.NullString{String: updateData.Location, Valid: true}
        }

        // Формирование SQL-запроса
        if len(updateFields) == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Нет данных для обновления"})
            return
        }

        query := "UPDATE users SET "
        params := []interface{}{}
        i := 1
        
        for field, value := range updateFields {
            query += fmt.Sprintf("%s = $%d, ", field, i)
            params = append(params, value)
            i++
        }
        
        query = strings.TrimSuffix(query, ", ")
        query += " WHERE id = $" + strconv.Itoa(i)
        params = append(params, userID)

        // Выполнение запроса
        _, err = db.Exec(query, params...)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении данных"})
            return
        }

        // Возврат обновленных данных
        updatedUser, err := database.GetUserByID(db, userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении обновленных данных"})
            return
        }

        response := gin.H{
            "id":         updatedUser.ID,
            "username":   updatedUser.Username,
        }
        
        if updatedUser.Email.Valid {
            response["email"] = updatedUser.Email.String
        }
        if updatedUser.FullName.Valid {
            response["full_name"] = updatedUser.FullName.String
        }
        if updatedUser.AvatarURL.Valid {
            response["avatar_url"] = updatedUser.AvatarURL.String
        }
        if updatedUser.PhoneNumber.Valid {
            response["phone_number"] = updatedUser.PhoneNumber.String
        }
        if updatedUser.DateOfBirth.Valid {
            response["date_of_birth"] = updatedUser.DateOfBirth.String
        }
        if updatedUser.Location.Valid {
            response["location"] = updatedUser.Location.String
        }

        c.JSON(http.StatusOK, response)
    }
}

func DeleteUserHandler(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Получаем ID пользователя через единую функцию
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // 2. Проверяем существование пользователя без учета deleted_at
        exists, err := database.UserExistsRaw(db, userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Ошибка при проверке пользователя",
                "details": err.Error(),
            })
            return
        }

        if !exists {
            c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
            return
        }

        // 3. Выполняем SOFT DELETE
        result, err := db.Exec(`
            UPDATE users 
            SET 
                deleted_at = NOW(),
                email = NULL,
                phone_number = NULL,
                username = CONCAT(username, '_deleted_', EXTRACT(EPOCH FROM NOW()))
            WHERE id = $1`, userID)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Ошибка при удалении пользователя",
                "details": err.Error(),
            })
            return
        }

        rowsAffected, _ := result.RowsAffected()
        if rowsAffected == 0 {
            c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже был удален"})
            return
        }

        // 4. Успешный ответ
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "Аккаунт успешно деактивирован",
            "user_id": userID,
            "deleted_at": time.Now().Format(time.RFC3339),
        })
    }
}

func GetUserResultsHandler(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        results, err := database.GetUserResults(db, userID)
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

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем токен из заголовка
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        // Удаляем префикс "Bearer " если есть
        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        }

        // Извлекаем ID пользователя из токена
        userID, err := jwt.ExtractUserIDFromJWT(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
            return
        }

        // Сохраняем ID пользователя в контекст
        c.Set("userID", userID)
        c.Next()
    }
}