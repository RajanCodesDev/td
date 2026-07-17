package task

import "database/sql"

type Task struct {
	ID   int
	Task string
}

func Add(db *sql.DB, text string) error {
	_, err := db.Exec(
		"INSERT INTO tasks(task) VALUES (?)",
		text,
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