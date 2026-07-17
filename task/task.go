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
	DueDate     *time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
	Recurring   string
	NextDue     *time.Time
}

func Add(db *sql.DB, text string) error {
	return AddTask(
		db,
		text,
		2,
		"",
		nil,
		"",
	)
}




func AddTask(
	db *sql.DB,
	text string,
	priority int,
	tags string,
	due *time.Time,
	recurring string,
) error {
	var dueString any
	var nextDue any
	if due != nil {
		dueString = due.Format(time.RFC3339)
	}

	if recurring != "" && due != nil {
		n := nextOccurrence(
			*due,
			recurring,
		)

		nextDue =
			n.Format(
				time.RFC3339,
			)
	}
	

	_, err := db.Exec(
		`
		INSERT INTO tasks (
			task,
			priority,
			tags,
			due_date,
			recurring,
			next_due,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		text,
		priority,
		tags,
		dueString,
		recurring,
		nextDue,
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

func Get(
	db *sql.DB,
	id int,
) (*Task, error) {
	rows, err := db.Query(`
		SELECT
			id,
			task,
			completed,
			priority,
			tags,
			due_date,
			recurring,
			next_due,
			created_at,
			completed_at
		FROM tasks
		WHERE
			task LIKE ?
			OR tags LIKE ?
		ORDER BY completed ASC,
				priority DESC,
				id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks, err :=
		scanTasks(rows)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, sql.ErrNoRows
	}

	return &tasks[0], nil
}


func nextOccurrence(
	t time.Time,
	r string,
) time.Time {
	switch r {
	case "daily":
		return t.AddDate(0, 0, 1)

	case "weekly":
		return t.AddDate(0, 0, 7)

	case "monthly":
		return t.AddDate(0, 1, 0)

	case "yearly":
		return t.AddDate(1, 0, 0)

	default:
		return t
	}
}


func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task

	var recurring sql.NullString
	var nextDue sql.NullString

	for rows.Next() {
		var t Task

		var created string
		var completedAt sql.NullString
		var tags sql.NullString
		var due sql.NullString

		err := rows.Scan(
				&t.ID,
				&t.Task,
				&t.Completed,
				&t.Priority,
				&tags,
				&due,
				&recurring,
				&nextDue,
				&created,
				&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if tags.Valid {
			t.Tags = tags.String
		}

		if recurring.Valid {
			t.Recurring =
				recurring.String
		}

		if nextDue.Valid {
			n, _ :=
				time.Parse(
					time.RFC3339,
					nextDue.String,
				)

			t.NextDue = &n
		}

		if due.Valid {
			d, _ :=
				time.Parse(
					time.RFC3339,
					due.String,
				)

			t.DueDate = &d
		}

		t.CreatedAt, _ =
			time.Parse(
				time.RFC3339,
				created,
			)

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

func List(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`
			SELECT
				id,
				task,
				completed,
				priority,
				tags,
				due_date,
				recurring,
				next_due,
				created_at,
				completed_at
			FROM tasks
		ORDER BY completed ASC,
		         priority DESC,
		         id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func Search(
	db *sql.DB,
	query string,
) ([]Task, error) {
	rows, err := db.Query(`
		SELECT
			id,
			task,
			completed,
			priority,
			tags,
			due_date,
			created_at,
			completed_at
		FROM tasks
		WHERE
			task LIKE ?
			OR tags LIKE ?
		ORDER BY completed ASC,
		         priority DESC,
		         id ASC
	`,
		"%"+query+"%",
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func Today(db *sql.DB) ([]Task, error) {
	tasks, err := List(db)
	if err != nil {
		return nil, err
	}

	var result []Task
	now := time.Now()

	for _, t := range tasks {
		if t.DueDate == nil {
			continue
		}

		if sameDay(*t.DueDate, now) {
			result = append(result, t)
		}
	}

	return result, nil
}

func Overdue(db *sql.DB) ([]Task, error) {
	tasks, err := List(db)
	if err != nil {
		return nil, err
	}

	var result []Task
	now := time.Now()

	for _, t := range tasks {
		if t.Completed {
			continue
		}

		if t.DueDate == nil {
			continue
		}

		if t.DueDate.Before(now) {
			result = append(result, t)
		}
	}

	return result, nil
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()

	return y1 == y2 &&
		m1 == m2 &&
		d1 == d2
}

func Done(
	db *sql.DB,
	id int,
) error {
	t, err := Get(db, id)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`
		UPDATE tasks
		SET completed = 1,
		    completed_at = ?
		WHERE id = ?
		`,
		time.Now().Format(time.RFC3339),
		id,
	)
	if err != nil {
		return err
	}

	if t.Recurring != "" &&
		t.DueDate != nil {

		next :=
			nextOccurrence(
				*t.DueDate,
				t.Recurring,
			)

		err =
			AddTask(
				db,
				t.Task,
				t.Priority,
				t.Tags,
				&next,
				t.Recurring,
			)
		if err != nil {
			return err
		}
	}

	return nil
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