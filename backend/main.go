package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Структуры для запроса к Ollama
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
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	ollamaURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434" // Для локального запуска
	}
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "gemma3:4b-it-qat"
	}

	r := gin.Default()
	r.Use(cors.Default())

	// Проверка здоровья
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "model": ollamaModel})
	})

	// Чат с ИИ
	r.POST("/api/chat", func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
			Role    string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Системный промпт: объясняем боту, кто он и как себя вести
		systemPrompt := `Ты — AI-buddy, ИИ-наставник для стажёров компании Авито. 
Твоя задача — помогать стажёру адаптироваться. Отвечай кратко, дружелюбно и по делу.
Если тебя спрашивают про отгул или доступы, скажи, что можешь помочь оформить заявку.
Не придумывай несуществующие правила компании.`

		// Формируем запрос к Ollama
		ollamaReq := OllamaRequest{
			Model: ollamaModel,
			Messages: []Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: req.Message},
			},
			Stream: false,
		}

		jsonData, _ := json.Marshal(ollamaReq)

		// Таймаут 2 минуты (локальные модели думают медленно)
		client := &http.Client{Timeout: 2 * time.Minute}
		resp, err := client.Post(ollamaURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не могу достучаться до Ollama: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var ollamaResp OllamaResponse
		if err := json.Unmarshal(body, &ollamaResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга ответа от Ollama"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"reply": ollamaResp.Message.Content})
	})

	log.Printf("Сервер запущен на порту %s. Модель: %s", port, ollamaModel)
	r.Run(":" + port)
}
