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

	r.GET("/api/progress", func(c *gin.Context) {
		var total, done int
		db.QueryRow("SELECT COUNT(*) FROM tasks WHERE LOWER(period) LIKE '%недел%'").Scan(&total)
		db.QueryRow("SELECT COUNT(*) FROM tasks WHERE LOWER(period) LIKE '%недел%' AND is_done = TRUE").Scan(&done)

		percent := 0.0
		if total > 0 {
			percent = float64(done) / float64(total) * 100
		}
		c.JSON(http.StatusOK, gin.H{"total": total, "done": done, "percent": percent})
	})

	r.GET("/api/tasks", func(c *gin.Context) {
		periods := []string{"1 неделя", "2 неделя", "1 месяц", "2 месяц", "3 месяц"}
		var resultTasks []map[string]interface{}
		currentPeriod := ""

		for _, p := range periods {
			rows, err := db.Query(`SELECT id, title, description, is_done FROM tasks 
			          WHERE LOWER(period) LIKE LOWER($1) 
			          AND LOWER(title) NOT LIKE 'добавь%' 
			          AND LOWER(title) NOT LIKE 'проверь%' 
			          AND LOWER(title) NOT LIKE 'назначь%' 
			          AND LOWER(title) NOT LIKE 'запроси%'
			          AND LOWER(title) NOT LIKE 'работа с наставником%'
			          AND is_done = FALSE`, "%"+p+"%")

			if err == nil {
				for rows.Next() {
					var id int
					var title, desc string
					var isDone bool
					rows.Scan(&id, &title, &desc, &isDone)
					resultTasks = append(resultTasks, map[string]interface{}{
						"id": id, "title": title, "description": desc, "is_done": isDone,
					})
				}
				rows.Close()
			}

			if len(resultTasks) > 0 {
				currentPeriod = p
				break
			}
		}

		if len(resultTasks) == 0 {
			c.JSON(200, []interface{}{})
			return
		}

		log.Printf("Отдаем задачи для периода: %s", currentPeriod)
		c.JSON(200, resultTasks)
	})

	r.POST("/api/tasks/:id/complete", func(c *gin.Context) {
		id := c.Param("id")
		_, err := db.Exec("UPDATE tasks SET is_done = TRUE WHERE id = $1", id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось обновить задачу"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/chat", func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
			Role    string `json:"role"`
			Name    string `json:"name"`
		}
		c.ShouldBindJSON(&req)

		userName := req.Name
		if userName == "" {
			userName = "Стажёр"
		}

		if strings.Contains(strings.ToLower(req.Message), "отгул") || strings.Contains(strings.ToLower(req.Message), "отпуск") {
			absID := fmt.Sprintf("ABS-%d", time.Now().Unix()%10000)
			reply := fmt.Sprintf("📅 %s, я оформил для тебя заявку на отсутствие.\n\n*Номер заявки:* `%s`\n*Статус:* На согласовании у руководителя.\n\n**Следующий шаг:** Отследить статус можно в сервисе отсутствий.", userName, absID)
			c.JSON(200, gin.H{"reply": reply})
			return
		}

		if strings.Contains(strings.ToLower(req.Message), "доступ") || strings.Contains(strings.ToLower(req.Message), "jira") || strings.Contains(strings.ToLower(req.Message), "confluence") {
			accID := fmt.Sprintf("ACC-%d", time.Now().Unix()%10000)
			reply := fmt.Sprintf("🔐 %s, я создал заявку на доступ.\n\n*Номер заявки:* `%s`\n*Ресурс:* Jira / Confluence\n*Уровень прав:* Чтение\n*Статус:* Направлена владельцу проекта.\n\n**Следующий шаг:** Заявка появится в личном кабинете.", userName, accID)
			c.JSON(200, gin.H{"reply": reply})
			return
		}

		// Если просят задачи — сразу отдаем из базы, не спрашивая LLM
		if strings.Contains(strings.ToLower(req.Message), "задачи") || strings.Contains(strings.ToLower(req.Message), "задач") {
			periods := []string{"1 неделя", "2 неделя", "1 месяц", "2 месяц", "3 месяц"}
			var tasks []string
			foundPeriod := ""

			for _, p := range periods {
				query := `SELECT title, description FROM tasks 
				          WHERE LOWER(period) LIKE LOWER($1) 
				          AND LOWER(title) NOT LIKE 'добавь%' 
				          AND LOWER(title) NOT LIKE 'проверь%' 
				          AND LOWER(title) NOT LIKE 'назначь%' 
				          AND LOWER(title) NOT LIKE 'запроси%'
				          AND LOWER(title) NOT LIKE 'работа с наставником%'
				          AND is_done = FALSE`

				rows, err := db.Query(query, "%"+p+"%")
				if err == nil {
					for rows.Next() {
						var title, desc string
						rows.Scan(&title, &desc)
						tasks = append(tasks, fmt.Sprintf("- %s: %s", title, desc))
					}
					rows.Close()
				}
				if len(tasks) > 0 {
					foundPeriod = p
					break
				}
			}

			if len(tasks) > 0 {
				var rawTasks strings.Builder
				for _, t := range tasks {
					rawTasks.WriteString(t + "\n")
				}
				beautifyPrompt := fmt.Sprintf("Ты — дружелюбный ИИ-наставник Авито. Вот список актуальных задач стажёра на период \"%s\":\n%s\nНапиши приветственное сообщение и перечисли эти задачи в виде красивого нумерованного списка. Обращайся на 'ты'.", foundPeriod, rawTasks.String())

				beautifyReq := OllamaRequest{
					Model:    ollamaModel,
					Messages: []Message{{Role: "user", Content: beautifyPrompt}},
					Stream:   false,
				}
				beautifyJson, _ := json.Marshal(beautifyReq)
				client := &http.Client{Timeout: 2 * time.Minute}
				beautifyResp, err := client.Post(ollamaURL+"/api/chat", "application/json", bytes.NewBuffer(beautifyJson))

				if err == nil {
					defer beautifyResp.Body.Close()
					body2, _ := io.ReadAll(beautifyResp.Body)
					var finalResp OllamaResponse
					json.Unmarshal(body2, &finalResp)
					if finalResp.Message.Content != "" {
						replyText := finalResp.Message.Content
						replyText += "\n\n---\n**Следующий шаг:** Выбери задачу в трекере справа и отметь её выполненной, когда закончишь."
						c.JSON(200, gin.H{"reply": replyText})
						return
					}
				}
				c.JSON(200, gin.H{"reply": "Вот твои задачи на " + foundPeriod + ":\n\n" + rawTasks.String()})
				return
			}

			c.JSON(200, gin.H{"reply": "Поздравляю! Ты выполнил все задачи на всех периодах. Ты молодец!"})
			return
		}

		// На все остальные вопросы отвечает LLM
		systemPrompt := fmt.Sprintf(`Ты — AI-buddy, дружелюбный ИИ-наставник стажёров Авито. 
Ты общаешься со стажёром по имени %s (роль: %s). Обращайся к нему по имени.
Отвечай кратко, вежливо и по делу.`, userName, req.Role)

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
			c.JSON(200, gin.H{"reply": "Извини, я временно не могу связаться с нейросетью. Попробуй позже."})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var ollamaResp OllamaResponse
		json.Unmarshal(body, &ollamaResp)

		c.JSON(200, gin.H{"reply": ollamaResp.Message.Content})
	})

	log.Printf("Сервер запущен. Модель: %s", ollamaModel)
	r.Run(":" + port)
}
