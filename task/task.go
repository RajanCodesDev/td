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
	Project     string
	DueDate     *time.Time
	CreatedAt   time.Time
	CompletedAt *time.Time
	Recurring   string
	NextDue     *time.Time
}

type ProjectCount struct {
	Name  string
	Count int
}


func ListProjects(db *sql.DB) ([]ProjectCount, error) {
	rows, err := db.Query(`
		SELECT
			COALESCE(NULLIF(TRIM(project),''),'default'),
			COUNT(*)
		FROM tasks
		GROUP BY COALESCE(NULLIF(TRIM(project),''),'default')
		ORDER BY
			CASE
				WHEN COALESCE(NULLIF(TRIM(project),''),'default')='default'
				THEN 0
				ELSE 1
			END,
			COALESCE(NULLIF(TRIM(project),''),'default')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectCount

	for rows.Next() {
		var p ProjectCount

		if err := rows.Scan(&p.Name, &p.Count); err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	return projects, nil
}

func ProjectTasks(
	db *sql.DB,
	project string,
) ([]Task, error) {

	var rows *sql.Rows
	var err error

	if project == "default" {
		rows, err = db.Query(`
			SELECT
				id,
				task,
				completed,
				priority,
				project,
				due_date,
				recurring,
				next_due,
				created_at,
				completed_at
			FROM tasks
			WHERE project IS NULL
			   OR project=''
		`)
	} else {
		rows, err = db.Query(`
			SELECT
				id,
				task,
				completed,
				priority,
				project,
				due_date,
				recurring,
				next_due,
				created_at,
				completed_at
			FROM tasks
			WHERE LOWER(project)=LOWER(?)
		`, project)
	}

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

		if startOfDay(*t.DueDate).
			Before(startOfDay(now)) {
			s.Overdue++
		}
	}

	return s, nil
}

func sortWeight(t Task) int {
	now := startOfDay(time.Now())

	if t.Completed {
		return 6
	}

	if t.DueDate == nil {
		return 5
	}

	due := startOfDay(*t.DueDate)

	switch {
	case due.Before(now):
		return 1

	case sameDay(due, now):
		return 2

	case sameDay(
		due,
		now.AddDate(0, 0, 1),
	):
		return 3

	default:
		return 4
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		t.Location(),
	)
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		a := tasks[i]
		b := tasks[j]

		wa := sortWeight(a)
		wb := sortWeight(b)

		if wa != wb {
			return wa < wb
		}

		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}

		if a.DueDate != nil &&
			b.DueDate != nil {

			da := startOfDay(*a.DueDate)
			db := startOfDay(*b.DueDate)

			if !da.Equal(db) {
				return da.Before(db)
			}
		}

		return a.ID < b.ID
	})
}

func AddTask(
	db *sql.DB,
	text string,
	priority int,
	project string,
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
			project,
			due_date,
			recurring,
			next_due,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		text,
		priority,
		project,
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
			project,
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
		var project sql.NullString
		var due sql.NullString

		err := rows.Scan(
			&t.ID,
			&t.Task,
			&t.Completed,
			&t.Priority,
			&project,
			&due,
			&recurring,
			&nextDue,
			&created,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if project.Valid {
			t.Project = project.String
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
			project,
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
			project,
			due_date,
			recurring,
			next_due,
			created_at,
			completed_at
		FROM tasks
		WHERE
			task LIKE ?
			OR project LIKE ?
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
		if t.Completed {
			continue
		}

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

	res, err := db.Exec(
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

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
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
				t.Project,
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
