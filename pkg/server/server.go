package server

import (
	"context"
	"errors"
	"fmt"
	"go1f/pkg/api"
	"go1f/pkg/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() error {
	pass := os.Getenv("TODO_PASSWORD")
	if pass != "" {
		log.Printf("Пароль установлен")
	} else {
		log.Println("Пароль не установлен")
	}

	log.Println("Инициализация базы данных...")
	if err := db.Init("scheduler.db"); err != nil {
		return fmt.Errorf("не удалось инициализировать БД: %w", err)
	}
	defer db.DB.Close()

	log.Println("База данных успешно инициализирована")

	port := 7540

	mux := http.NewServeMux()
	if _, err := os.Stat("web"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("web")))
		log.Println("Web-интерфейс подключен")
	} else {
		log.Println("Папка web не найдена, работаем только как API")
	}

	api.Init(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("listen and serve: %w", err)
	case sig := <-shutdown:
		log.Printf("получен сигнал %s, выключаемся", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}

	return nil
}
