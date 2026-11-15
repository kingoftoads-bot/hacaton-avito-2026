package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

// Структура задачи
type Task struct {
	Period string `json:"period"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
}

// Подключение к базе и создание таблицы
func InitDB() *sql.DB {
	// Жестко прописываем строку подключения (пароль root)
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

	// Создаем таблицу с колонкой is_done
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

	// Если база уже была создана ранее (без колонки is_done), добавляем ее
	db.Exec("ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_done BOOLEAN DEFAULT FALSE")

	return db
}

// Чтение Excel и загрузка в базу
func SeedDB(db *sql.DB) {
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
			continue // Пропускаем заголовки
		}
		if len(row) >= 3 {
			_, err := db.Exec("INSERT INTO tasks (title, description, period) VALUES ($1, $2, $3)", row[0], row[1], row[2])
			if err == nil {
				inserted++
			}
		}
	}
	fmt.Printf("✅ Загружено %d задач из Excel в PostgreSQL\n", inserted)
}
