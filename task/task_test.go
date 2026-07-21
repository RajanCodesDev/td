package task

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func testDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 0,
		priority INTEGER NOT NULL DEFAULT 2,
		project TEXT,
		due_date TEXT,
		recurring TEXT,
		next_due TEXT,
		created_at TEXT NOT NULL,
		completed_at TEXT
	)
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}


func TestAddAndGet(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	err := Add(db, "hello")
	if err != nil {
		t.Fatal(err)
	}

	task, err := Get(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if task.Task != "hello" {
		t.Fatalf(
			"expected hello got %s",
			task.Task,
		)
	}
}


func TestDelete(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	Add(db, "hello")

	err := Delete(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Get(db, 1)
	if err != sql.ErrNoRows {
		t.Fatal("expected no rows")
	}
}

func TestDeleteMissing(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	err := Delete(db, 999)
	if err != sql.ErrNoRows {
		t.Fatal("expected sql.ErrNoRows")
	}
}

func TestModify(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	Add(db, "hello")

	err := Modify(
		db,
		1,
		"world",
	)
	if err != nil {
		t.Fatal(err)
	}

	task, _ := Get(db, 1)

	if task.Task != "world" {
		t.Fatal("modify failed")
	}
}

func TestDoneUndo(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	Add(db, "hello")

	err := Done(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	task, _ := Get(db, 1)

	if !task.Completed {
		t.Fatal("should be completed")
	}

	err = Undo(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	task, _ = Get(db, 1)

	if task.Completed {
		t.Fatal("should be pending")
	}
}

func TestToday(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	due := time.Now()

	err := AddTask(
		db,
		"today",
		2,
		"",
		&due,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := Today(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("expected 1 task")
	}
}

func TestOverdue(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	due :=
		time.Now().
			AddDate(0, 0, -1)

	AddTask(
		db,
		"overdue",
		2,
		"",
		&due,
		"",
	)

	tasks, err := Overdue(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatal("expected overdue task")
	}
}

func TestRecurringTask(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	due := time.Now()

	AddTask(
		db,
		"backup",
		2,
		"",
		&due,
		"daily",
	)

	err := Done(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := List(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 2 {
		t.Fatalf(
			"expected 2 tasks got %d",
			len(tasks),
		)
	}
}