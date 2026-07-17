package task

import "database/sql"

type Task struct {
	ID          int
	Task        string
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func Add(db *sql.DB, text string) error {
	_, err := db.Exec(
		`INSERT INTO tasks (
			task,
			created_at
		)
		VALUES (?, ?)`,
		text,
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
	rows, err := db.Query(
		"SELECT id, task FROM tasks ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var t Task

		err := rows.Scan(&t.ID, &t.Task)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}


func Done(db *sql.DB, id int) error
func Undo(db *sql.DB, id int) error