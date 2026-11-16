package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

type Task struct {
	Period string `json:"period"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
}

func InitDB() *sql.DB {
	connStr := "user=postgres password=root dbname=aibuddy host=localhost port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("❌ ОШИБКА БАЗЫ: %v\nПроверь: запущен ли Postgres, верный ли пароль 'root', создана ли база 'aibuddy'", err)
	}
	fmt.Println("✅ Успешно подключились к PostgreSQL")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		period TEXT,
		title TEXT,
		description TEXT,
		is_done BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}

	return db
}

func SeedDB(db *sql.DB) {
	// Сбрасываем все галочки при запуске
	db.Exec("UPDATE tasks SET is_done = FALSE")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if count > 0 {
		fmt.Println("✅ Задачи уже есть в базе, пропускаем парсинг Excel")
		return
	}

	f, err := excelize.OpenFile("checklist.xlsx")
	if err != nil {
		log.Println("⚠️ Ошибка чтения Excel:", err)
		return
	}
	defer f.Close()

	rows, _ := f.GetRows("Материалы для всех")
	inserted := 0
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) >= 3 {
			title := row[0]
			desc := row[1]
			period := row[2]

			replacer := strings.NewReplacer(
				"Добавь стажёра", "Меня добавили",
				"Проверь, запросил ли стажёр", "Я запросил",
				"Назначь со стажёром", "У меня прошли встречи",
				"Обсуди цели", "Обсудили цели",
				"Запроси у куратора", "Запросил у куратора",
				"Попроси добавить себя", "Попросил добавить меня",
				"Обсуди с наставником", "Обсудил с наставником",
				"Заведи конспект", "Завел конспект",
				"Запроси доступ", "Запросил доступ",
			)
			desc = replacer.Replace(desc)

			if strings.Contains(desc, "1.") || strings.Contains(desc, "2.") {
				parts := regexp.MustCompile(`\d+\.\s`).Split(desc, -1)
				for _, part := range parts {
					cleanPart := strings.TrimSpace(part)
					if len(cleanPart) > 10 {
						_, err := db.Exec("INSERT INTO tasks (title, description, period) VALUES ($1, $2, $3)", title, cleanPart, period)
						if err == nil {
							inserted++
						}
					}
				}
			} else {
				_, err := db.Exec("INSERT INTO tasks (title, description, period) VALUES ($1, $2, $3)", title, desc, period)
				if err == nil {
					inserted++
				}
			}
		}
	}
	fmt.Printf("✅ Загружено %d задач из Excel в PostgreSQL\n", inserted)
}
