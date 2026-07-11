package db

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT DEFAULT '',
    repeat TEXT DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

func Init(dbFile string) error {
	log.Printf("Init: открываем БД %s", dbFile)

	_, err := os.Stat(dbFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		log.Printf("Init: ошибка открытия: %v", err)
		return err
	}

	if err = DB.Ping(); err != nil {
		log.Printf("Init: ошибка Ping: %v", err)
		DB.Close()
		return err
	}

	log.Println("Init: создаем таблицу...")
	if _, err = DB.Exec(schema); err != nil {
		log.Printf("Init: ошибка создания таблицы: %v", err)
		DB.Close()
		return err
	}

	log.Printf("Init: БД успешно инициализирована")
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
