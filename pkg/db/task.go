package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	queryDeleteTask = `DELETE FROM scheduler WHERE id = ?`
	queryUpdateDate = `UPDATE scheduler SET date = ? WHERE id = ?`
	queryGetTask    = `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	queryUpdateTask = `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	queryTask       = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
	queryAddTask    = `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func DeleteTask(ctx context.Context, id string) error {
	_, err := DB.ExecContext(ctx, queryDeleteTask, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу: %w", err)
	}
	return nil
}

func UpdateDate(ctx context.Context, nextDate string, id string) error {
	_, err := DB.ExecContext(ctx, queryUpdateDate, nextDate, id)
	if err != nil {
		return fmt.Errorf("Не удалось обновить дату задачи: %w", err)
	}
	return nil
}
func GetTask(ctx context.Context, id string) (*Task, error) {
	var t Task
	err := DB.QueryRowContext(ctx, queryGetTask, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Задача не найдена: %w", err)
		}
		return nil, err
	}
	return &t, nil
}

func UpdateTask(ctx context.Context, task *Task) error {
	res, err := DB.ExecContext(ctx, queryUpdateTask, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
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
	tasks := []Task{}
	rows, err := DB.QueryContext(ctx, queryTask, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func AddTask(ctx context.Context, task *Task) (int64, error) {
	res, err := DB.ExecContext(ctx, queryAddTask, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
