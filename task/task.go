package task

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          int
	Task        string
	Completed   bool
	Priority    int
	Tags        string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func Add(db *sql.DB, text string) error {
	return AddWithPriority(db, text, 2)
}

func AddWithPriority(
	db *sql.DB,
	text string,
	priority int,
) error {
	_, err := db.Exec(
		`
		INSERT INTO tasks (
			task,
			priority,
			created_at
		)
		VALUES (?, ?, ?)
		`,
		text,
		priority,
		time.Now().Format(time.RFC3339),
	)

	return err
}

func Delete(db *sql.DB, id int) error {
	_, err := db.Exec(
		"DELETE FROM tasks WHERE id=?",
		id,
	)

	return err
}

func Modify(db *sql.DB, id int, text string) error {
	_, err := db.Exec(
		"UPDATE tasks SET task=? WHERE id=?",
		text,
		id,
	)

	return err
}

func List(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`
		SELECT
			id,
			task,
			completed,
			priority,
			created_at,
			completed_at
		FROM tasks
		ORDER BY completed ASC, priority DESC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task

		var created string
		var completedAt sql.NullString

		err := rows.Scan(
			&t.ID,
			&t.Task,
			&t.Completed,
			&t.Priority,
			&created,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt, _ =
			time.Parse(time.RFC3339, created)

		if completedAt.Valid {
			c, _ :=
				time.Parse(
					time.RFC3339,
					completedAt.String,
				)

			t.CompletedAt = &c
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func Done(db *sql.DB, id int) error {
	_, err := db.Exec(
		`
		UPDATE tasks
		SET completed = 1,
		    completed_at = ?
		WHERE id = ?
		`,
		time.Now().Format(time.RFC3339),
		id,
	)

	return err
}

func Undo(db *sql.DB, id int) error {
	_, err := db.Exec(
		`
		UPDATE tasks
		SET completed = 0,
		    completed_at = NULL
		WHERE id = ?
		`,
		id,
	)

	return err
}