package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ==========================================
// СТРУКТУРЫ
// ==========================================
type OllamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaResponse struct {
	Message Message `json:"message"`
}

func main() {
	db := InitDB()
	SeedDB(db)
	defer db.Close()

	port := "8080"
	ollamaURL := "http://localhost:11434"
	ollamaModel := "gemma3:4b-it-qat"

	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/api/chat", func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
		}
		c.ShouldBindJSON(&req)

		// 1. УМНЫЙ ПРОМПТ (ReAct-подход)
		// Мы объясняем модели правила игры текстом, а не через Tools API
		systemPrompt := `Ты — AI-buddy, ИИ-наставник стажёров Авито. 
Если пользователь спрашивает про задачи, доступы или отгул, ты должен понять, какой период его интересует.
Если тебе нужны задачи из базы, напиши СТРОГО такую команду: [GET_TASKS: <период>]
Где <период> — это "1 неделя", "1 месяц" или "3 месяц".

Пример твоего ответа на вопрос "что делать на первой неделе?":
[GET_TASKS: 1 неделя]

Пример ответа на "привет":
Привет! Я твой ИИ-наставник. Спроси меня о задачах на неделю.`

		// 2. Первый запрос к модели (чтобы она поняла интент)
		ollamaReq := OllamaRequest{
			Model: ollamaModel,
			Messages: []Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: req.Message},
			},
			Stream: false,
		}

		jsonData, _ := json.Marshal(ollamaReq)
		client := &http.Client{Timeout: 2 * time.Minute}

		resp, err := client.Post(ollamaURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(500, gin.H{"error": "Ollama недоступна: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var ollamaResp OllamaResponse
		json.Unmarshal(body, &ollamaResp)

		aiAnswer := ollamaResp.Message.Content

		// 3. Проверяем: если модель написала [GET_TASKS: ...], мы сами лезем в базу
		if strings.Contains(aiAnswer, "[GET_TASKS:") {
			// Вытаскиваем период из скобок
			startIdx := strings.Index(aiAnswer, "[GET_TASKS:") + len("[GET_TASKS:")
			endIdx := strings.Index(aiAnswer[startIdx:], "]")
			period := ""
			if endIdx != -1 {
				period = strings.TrimSpace(aiAnswer[startIdx : startIdx+endIdx])
			}

			// Идем в PostgreSQL
			rows, err := db.Query("SELECT title, description FROM tasks WHERE LOWER(period) LIKE LOWER($1)", "%"+period+"%")
			if err == nil {
				var result strings.Builder
				result.WriteString(fmt.Sprintf("Вот твои задачи на **%s**:\n\n", period))
				count := 0
				for rows.Next() {
					var title, desc string
					rows.Scan(&title, &desc)
					result.WriteString(fmt.Sprintf("📌 **%s**\n%s\n\n", title, desc))
					count++
				}
				rows.Close()

				if count > 0 {
					c.JSON(200, gin.H{"reply": result.String()})
					return
				}
			}
		}

		// 4. Если команды не было — возвращаем обычный ответ
		c.JSON(200, gin.H{"reply": aiAnswer})
	})

	log.Printf("Сервер запущен. Модель: %s", ollamaModel)
	r.Run(":" + port)
}
