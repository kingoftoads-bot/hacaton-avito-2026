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
			Role    string `json:"role"`
		}
		c.ShouldBindJSON(&req)

		if strings.Contains(strings.ToLower(req.Message), "отгул") || strings.Contains(strings.ToLower(req.Message), "отпуск") {
			// Имитируем вызов корпоративного сервиса
			absID := fmt.Sprintf("ABS-%d", time.Now().Unix()%10000)
			reply := fmt.Sprintf("📅 Я оформил для тебя заявку на отсутствие.\n\n**Номер заявки:** `%s`\n**Статус:** На согласовании у руководителя.\n\nЯ пришлю уведомление, когда заявку одобрят.", absID)
			c.JSON(200, gin.H{"reply": reply})
			return
		}

		if strings.Contains(strings.ToLower(req.Message), "доступ") || strings.Contains(strings.ToLower(req.Message), "jira") || strings.Contains(strings.ToLower(req.Message), "confluence") {
			// Имитируем вызов сервиса доступов
			accID := fmt.Sprintf("ACC-%d", time.Now().Unix()%10000)
			reply := fmt.Sprintf("🔐 Я создал заявку на доступ.\n\n**Номер заявки:** `%s`\n**Ресурс:** Jira / Confluence\n**Уровень прав:** Чтение\n\nЗаявка направлена владельцу проекта.", accID)
			c.JSON(200, gin.H{"reply": reply})
			return
		}

		systemPrompt := `Ты — AI-buddy, ИИ-наставник стажёров Авито. 
Если пользователь спрашивает про задачи, напиши СТРОГО такую команду: [GET_TASKS: <период>]
Где <период> — это "1 неделя", "1 месяц" или "3 месяц".
Если это просто приветствие, поздоровайся.`

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

		// Если модель запросила задачи
		if strings.Contains(aiAnswer, "[GET_TASKS:") {
			startIdx := strings.Index(aiAnswer, "[GET_TASKS:") + len("[GET_TASKS:")
			endIdx := strings.Index(aiAnswer[startIdx:], "]")
			period := "1 неделя"
			if endIdx != -1 {
				period = strings.TrimSpace(aiAnswer[startIdx : startIdx+endIdx])
			}

			// Фильтруем задачи для наставника (убираем "Добавь", "Проверь" и т.д.)
			query := `SELECT title, description FROM tasks 
			          WHERE LOWER(period) LIKE LOWER($1) 
			          AND LOWER(title) NOT LIKE 'добавь%' 
			          AND LOWER(title) NOT LIKE 'проверь%' 
			          AND LOWER(title) NOT LIKE 'назначь%' 
			          AND LOWER(title) NOT LIKE 'запроси%'`

			rows, err := db.Query(query, "%"+period+"%")
			if err == nil {
				var rawTasks strings.Builder
				count := 0
				for rows.Next() {
					var title, desc string
					rows.Scan(&title, &desc)
					rawTasks.WriteString(fmt.Sprintf("- %s: %s\n", title, desc))
					count++
				}
				rows.Close()

				if count > 0 {
					// Второй запрос к ИИ, чтобы красиво оформить
					beautifyPrompt := fmt.Sprintf("Ты — дружелюбный ИИ-наставник Авито. Вот список задач стажёра на период \"%s\":\n%s\nНапиши приветственное сообщение и перечисли эти задачи в виде красивого нумерованного списка. Обращайся на 'ты'.", period, rawTasks.String())
					beautifyReq := OllamaRequest{
						Model:    ollamaModel,
						Messages: []Message{{Role: "user", Content: beautifyPrompt}},
						Stream:   false,
					}
					beautifyJson, _ := json.Marshal(beautifyReq)
					beautifyResp, err := client.Post(ollamaURL+"/api/chat", "application/json", bytes.NewBuffer(beautifyJson))
					if err == nil {
						defer beautifyResp.Body.Close()
						body2, _ := io.ReadAll(beautifyResp.Body)
						var finalResp OllamaResponse
						json.Unmarshal(body2, &finalResp)
						if finalResp.Message.Content != "" {
							c.JSON(200, gin.H{"reply": finalResp.Message.Content})
							return
						}
					}
					c.JSON(200, gin.H{"reply": "Вот твои задачи на " + period + ":\n\n" + rawTasks.String()})
					return
				}
			}
			c.JSON(200, gin.H{"reply": "На этот период задач пока нет."})
			return
		}

		c.JSON(200, gin.H{"reply": aiAnswer})
	})

	log.Printf("Сервер запущен. Модель: %s", ollamaModel)
	r.Run(":" + port)
}
