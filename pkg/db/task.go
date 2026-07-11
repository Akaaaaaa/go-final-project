package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func DeleteTask(ctx context.Context, id string) error {
	if DB == nil {
		return fmt.Errorf("база данных не инициализирована")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	_, err := DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу: %w", err)
	}
	return nil
}

func UpdateDate(ctx context.Context, nextDate string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	_, err := DB.ExecContext(ctx, query, nextDate, id)
	if err != nil {
		return fmt.Errorf("Не удалось обновить дату задачи: %w", err)
	}
	return nil
}
func GetTask(ctx context.Context, id string) (*Task, error) {
	var t Task
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`

	err := DB.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Задача не найдена: %w", err) // Задача не найдена
		}
		return nil, err
	}
	return &t, nil
}

func UpdateTask(ctx context.Context, task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`

	res, err := DB.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("Не удалось обновить задачу: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Ошибка количесто затронутых строк: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}

	return nil
}

func Tasks(ctx context.Context, limit int) ([]Task, error) {
	log.Println("=== db.Tasks START ===")
	tasks := []Task{}

	if DB == nil {
		log.Println("ОШИБКА: DB == nil")
		return tasks, fmt.Errorf("база данных не инициализирована")
	}

	log.Println("Проверяем существование таблицы...")
	var tableName string
	err := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='scheduler'").Scan(&tableName)
	if err != nil {
		log.Printf("Таблица не найдена, ошибка: %v", err)
		log.Println("Создаем таблицу...")
		if _, err := DB.Exec(schema); err != nil {
			log.Printf("Ошибка создания таблицы: %v", err)
			return tasks, fmt.Errorf("не удалось создать таблицу: %w", err)
		}
		log.Println("Таблица создана")
	} else {
		log.Printf("Таблица найдена: %s", tableName)
	}

	log.Printf("Выполняем запрос с limit=%d", limit)
	rows, err := DB.QueryContext(ctx, "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?", limit)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		return tasks, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			return tasks, err
		}
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Ошибка rows.Err(): %v", err)
		return tasks, err
	}

	log.Printf("db.Tasks возвращает %d задач", len(tasks))
	return tasks, nil
}

func AddTask(ctx context.Context, task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.ExecContext(ctx, query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
