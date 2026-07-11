package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("нет env")
	}

	log.Println("Запуск сервера...")
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("не удалось инициализировать БД: %v", err)
	}
	defer db.DB.Close()

	log.Println("База данных успешно инициализирована")

	if err := server.Run(); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

}
