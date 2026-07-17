package task

import (
	"database/sql"
	"sort"
	"time"
)

type Stats struct {
	Total     int
	Pending   int
	Completed int
	Overdue   int
	Today     int
}

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

func GetStats(db *sql.DB) (Stats, error) {
	var s Stats

	tasks, err := List(db)
	if err != nil {
		return s, err
	}

	now := time.Now()

	for _, t := range tasks {
		s.Total++

		if t.Completed {
			s.Completed++
			continue
		}

		s.Pending++

		if t.DueDate == nil {
			continue
		}

		if sameDay(*t.DueDate, now) {
			s.Today++
		}

		if t.DueDate.Before(now) &&
			!sameDay(*t.DueDate, now) {
			s.Overdue++
		}
	}

	return s, nil
}

func sortWeight(t Task) int {
	now := time.Now()

	if t.Completed {
		return 6
	}

	if t.DueDate != nil {

		if t.DueDate.Before(now) &&
			!sameDay(*t.DueDate, now) {
			return 1
		}

		if sameDay(*t.DueDate, now) {
			return 2
		}

		if sameDay(
			*t.DueDate,
			now.AddDate(0, 0, 1),
		) {
			return 3
		}

		return 4
	}

	return 5
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks,
		func(i, j int) bool {
			a := sortWeight(tasks[i])
			b := sortWeight(tasks[j])

			if a != b {
				return a < b
			}

			if tasks[i].Priority != tasks[j].Priority {
				return tasks[i].Priority >
					tasks[j].Priority
			}

			if tasks[i].DueDate != nil &&
				tasks[j].DueDate != nil {

				return tasks[i].DueDate.Before(
					*tasks[j].DueDate,
				)
			}

			return tasks[i].ID <
				tasks[j].ID
		},
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
	res, err := db.Exec(
		"DELETE FROM tasks WHERE id=?",
		id,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func Modify(
	db *sql.DB,
	id int,
	text string,
) error {
	res, err := db.Exec(
		"UPDATE tasks SET task=? WHERE id=?",
		text,
		id,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
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
			WHERE id = ?
		`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	sortTasks(tasks)

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
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	sortTasks(tasks)

	return tasks, nil
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
			recurring,
			next_due,
			created_at,
			completed_at
		FROM tasks
		WHERE
			task LIKE ?
			OR tags LIKE ?
	`,
		"%"+query+"%",
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	sortTasks(tasks)

	return tasks, nil
}

func ClearCompleted(db *sql.DB) error {
	_, err := db.Exec(`
		DELETE FROM tasks
		WHERE completed = 1
	`)
	return err
}

func nextFutureOccurrence(
	t time.Time,
	r string,
) time.Time {
	now := time.Now()

	for !t.After(now) {
		t = nextOccurrence(t, r)
	}

	return t
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

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	for _, t := range tasks {
		if t.Completed {
			continue
		}

		if t.DueDate == nil {
			continue
		}

		if t.DueDate.Before(today) {
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
			nextFutureOccurrence(
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

func Undo(
	db *sql.DB,
	id int,
) error {
	res, err := db.Exec(
		`
		UPDATE tasks
		SET completed = 0,
		    completed_at = NULL
		WHERE id = ?
		`,
		id,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}
