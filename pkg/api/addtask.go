package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go1f/pkg/db"
)

func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("дата представлена в формате, отличном от 20060102")
	}

	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	if task.Repeat != "" {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("правило повторения указано в неправильном формате: %w", err)
		}

		if tDate.Before(nowDate) || tDate.Equal(nowDate) {
			task.Date = nextDate
		}
	} else {
		if tDate.Before(nowDate) {
			task.Date = now.Format(DateFormat)
		}
	}

	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
		return
	}

	if task.Title == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Поле 'title' обязательно для заполнения"})
		return
	}

	err = checkDate(&task)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(r.Context(), &task)
	if err != nil {
		log.Printf("Ошибка при добавлении задачи: %v", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при добавлении задачи"})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}
