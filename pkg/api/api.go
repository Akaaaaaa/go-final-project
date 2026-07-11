package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"go1f/pkg/db"
)

func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/signin", SigninHandler)
	mux.HandleFunc("/api/nextdate", NextDateHandler)

	mux.HandleFunc("/api/task", AuthMiddleware(TaskHandler))
	mux.HandleFunc("/api/tasks", AuthMiddleware(TasksHandler))
	mux.HandleFunc("/api/task/done", AuthMiddleware(TaskDoneHandler))
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasksHandler(w, r)
	case http.MethodPost:
		done := r.URL.Query().Get("done")
		if done != "" {
			handleDoneTask(w, r)
			return
		}
		addTaskHandler(w, r)
	default:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func handleDoneTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is required"})
		return
	}

	log.Printf("handleDoneTask: marking task %s as done", id)

	task, err := db.GetTask(r.Context(), id)
	if err != nil {
		log.Printf("handleDoneTask: задача не найдена: %v", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	if task.Repeat != "" {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("handleDoneTask: ошибка вычисления следующей даты: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка вычисления следующей даты"})
			return
		}

		log.Printf("handleDoneTask: обновляем дату с %s на %s", task.Date, nextDate)
		if err := db.UpdateDate(r.Context(), nextDate, id); err != nil {
			log.Printf("handleDoneTask: ошибка обновления даты: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	} else {
		log.Printf("handleDoneTask: удаляем задачу %s", id)
		if err := db.DeleteTask(r.Context(), id); err != nil {
			log.Printf("handleDoneTask: ошибка удаления: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=== getTasksHandler START ===")

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // значение по умолчанию
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if limit == 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string][]db.Task{"tasks": {}})
		return
	}

	tasks, err := db.Tasks(r.Context(), limit)
	if err != nil {
		log.Printf("Ошибка получения задач: %v", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if tasks == nil {
		tasks = []db.Task{}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]db.Task{"tasks": tasks})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is required"})
		return
	}

	task, err := db.GetTask(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
		return
	}

	if task.ID == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи обязателен"})
		return
	}

	if task.Title == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Поле 'title' обязательно для заполнения"})
		return
	}

	if err := checkDate(&task); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateTask(r.Context(), &task); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
}

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var idStr string
	var hasID bool

	idFromQuery := r.URL.Query().Get("id")
	if idFromQuery != "" {
		idStr = idFromQuery
		hasID = true
		log.Printf("TaskDoneHandler: id из URL параметра = %s", idStr)
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("TaskDoneHandler: ошибка чтения body: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
			return
		}
		defer r.Body.Close()

		log.Printf("TaskDoneHandler: body = %s", string(body))

		if len(body) == 0 {
			log.Printf("TaskDoneHandler: пустое тело запроса")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
			return
		}

		var req struct {
			ID json.RawMessage `json:"id"`
		}

		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("TaskDoneHandler: ошибка парсинга JSON: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
			return
		}

		if len(req.ID) == 0 || string(req.ID) == "null" {
			log.Printf("TaskDoneHandler: ID отсутствует в JSON")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи обязателен"})
			return
		}

		if err := json.Unmarshal(req.ID, &idStr); err != nil {
			var idNum int64
			if err := json.Unmarshal(req.ID, &idNum); err != nil {
				log.Printf("TaskDoneHandler: неверный формат ID: %s", string(req.ID))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат ID"})
				return
			}
			idStr = fmt.Sprintf("%d", idNum)
		}
		hasID = true
		log.Printf("TaskDoneHandler: id из JSON = %s", idStr)
	}

	if !hasID || idStr == "" {
		log.Printf("TaskDoneHandler: ID отсутствует")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID задачи обязателен"})
		return
	}

	task, err := db.GetTask(r.Context(), idStr)
	if err != nil {
		log.Printf("TaskDoneHandler: задача не найдена: %v", err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	log.Printf("TaskDoneHandler: задача найдена: %+v", task)

	if task.Repeat != "" {
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("TaskDoneHandler: ошибка вычисления следующей даты: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка вычисления следующей даты"})
			return
		}

		log.Printf("TaskDoneHandler: обновляем дату с %s на %s", task.Date, nextDate)
		if err := db.UpdateDate(r.Context(), nextDate, idStr); err != nil {
			log.Printf("TaskDoneHandler: ошибка обновления даты: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	} else {
		log.Printf("TaskDoneHandler: удаляем задачу %s", idStr)
		if err := db.DeleteTask(r.Context(), idStr); err != nil {
			log.Printf("TaskDoneHandler: ошибка удаления: %v", err)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is required"})
		return
	}

	_, err := db.GetTask(r.Context(), id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Задача не найдена"})
		return
	}

	if err := db.DeleteTask(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
