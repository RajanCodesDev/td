package main

import (
	"database/sql"
	"fmt"
	"gocli/db"
	"gocli/editor"
	"gocli/task"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	Version = "v1.4.0"
	Reset  = "\033[0m"
	Green  = "\033[32m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".local", "share", "td")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "tasks.db"), nil
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

func usage() {
	fmt.Print(`
td add "task"
td add "task" --priority high
td add "task" --tags work,infra
td add "task" --priority high --tags work,infra
td add "task" --due 2026-08-01
td today
td overdue
td search <text>
td add
td add -e
td list
td done <id>
td undo <id>
td modify <id> "new task"
td delete <id>
td path
`)
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()

	return y1 == y2 &&
		m1 == m2 &&
		d1 == d2
}

func renderTable(title string, tasks []task.Task) {
	if len(tasks) == 0 {
		return
	}

	fmt.Printf("\n%s (%d)\n", title, len(tasks))

	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)

	isCompleted :=
		title == "Completed"

	if isCompleted {
		tw.AppendHeader(table.Row{
			"ID",
			"S",
			"P",
			"Completed",
			"Tags",
			"Task",
		})
	} else {
		tw.AppendHeader(table.Row{
			"ID",
			"S",
			"P",
			"Due",
			"Every",
			"Tags",
			"Task",
		})
	}

for _, t := range tasks {
	status := "○"
	if t.Completed {
		status = "✓"
	}

	priority := "M"

	switch t.Priority {
	case 1:
		priority = "L"
	case 2:
		priority = "M"
	case 3:
		priority = "H"
	}

	tags := "-"
	if t.Tags != "" {
		tags = "[" + t.Tags + "]"
	}

	if isCompleted {
		completed := "-"

		if t.CompletedAt != nil {
			completed =
				t.CompletedAt.Format(
					"2006-01-02",
				)
		}

		tw.AppendRow(table.Row{
			t.ID,
			status,
			priority,
			completed,
			tags,
			t.Task,
		})

		continue
	}

	due := "-"

	if t.DueDate != nil {
		now := time.Now()
		tomorrow := now.AddDate(0, 0, 1)

		switch {
		case sameDay(*t.DueDate, now):
			due = "TODAY"

		case sameDay(*t.DueDate, tomorrow):
			due = "TOMORROW"

		case t.DueDate.Before(now):
			days :=
				int(
					startOfDay(now).
						Sub(startOfDay(*t.DueDate)).
						Hours() / 24,
				)	

			due = fmt.Sprintf(
				"OVERDUE (%dd)",
				days,
			)

		default:
			days :=
				int(
					startOfDay(*t.DueDate).
						Sub(startOfDay(now)).
						Hours() / 24,
				)

			if days <= 90 {
				due = fmt.Sprintf(
					"+%dd",
					days,
				)
			} else {
				due = t.DueDate.Format(
					"2006-01-02",
				)
			}
		}
	}

	every := "-"

	if t.Recurring != "" {
		every =
			"↻ " +
				t.Recurring
	}

	tw.AppendRow(table.Row{
		t.ID,
		status,
		priority,
		due,
		every,
		tags,
		t.Task,
	})
}

	tw.SetStyle(table.StyleRounded)

	tw.Style().Format.Header =
		text.FormatDefault

	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateColumns = true
	tw.Style().Options.SeparateHeader = true

	tw.Render()
}


func printTasks(tasks []task.Task) {
	var pending []task.Task
	var completed []task.Task

	for _, t := range tasks {
		if t.Completed {
			completed = append(completed, t)
		} else {
			pending = append(pending, t)
		}
	}

	renderTable("Pending", pending)
	renderTable("Completed", completed)
}


func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	path, err := dbPath()
	if err != nil {
		panic(err)
	}

	database, err := db.Init(path)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	command := os.Args[1]
	switch command {
		case "ls":
			command = "list"

		case "rm", "del":
			command = "delete"

		case "c":
			command = "done"

		case "u":
			command = "undo"

		case "t":
			command = "today"

		case "o":
			command = "overdue"

		case "s":
			command = "search"

		case "stat":
			command = "stats"

		case "cc":
			command = "clear-completed"

		case "v":
			command = "version"
		}

	switch command {

	case "clear-completed":
		err := task.ClearCompleted(database)
		if err != nil {
			panic(err)
		}

		fmt.Println("Completed tasks removed.")

	case "stats":
		s, err := task.GetStats(database)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Total      : %d\n", s.Total)
		fmt.Printf("Pending    : %d\n", s.Pending)
		fmt.Printf("Completed  : %d\n", s.Completed)
		fmt.Printf("Overdue    : %d\n", s.Overdue)
		fmt.Printf("Due Today  : %d\n", s.Today)

		if s.Total > 0 {
			p :=
				float64(s.Completed) /
					float64(s.Total) *
					100

			fmt.Printf(
				"Completion : %.1f%%\n",
				p,
			)
		}



	case "add":

		if len(os.Args) == 2 {
			tasks, err := editor.Open()
			if err != nil {
				panic(err)
			}

			for _, t := range tasks {
				err := task.Add(database, t)
				if err != nil {
					panic(err)
				}
			}

			fmt.Printf("Added %d tasks.\n", len(tasks))
			return
		}

		if os.Args[2] == "-e" {
			fmt.Print("Editor: ")

			var name string
			fmt.Scanln(&name)

			err := editor.SetEditor(name)
			if err != nil {
				panic(err)
			}

			fmt.Println("Editor updated.")
			return
		}

		priority := 2
		tags := ""
		var due *time.Time
		recurring := ""


		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--every" {
				if i+1 >= len(os.Args) {
					fmt.Println("missing recurrence")
					return
				}

				r :=
					strings.ToLower(
						os.Args[i+1],
					)

				switch r {
				case "daily",
					"weekly",
					"monthly",
					"yearly":

					recurring = r

				default:
					fmt.Println(
						"recurrence must be daily, weekly, monthly or yearly",
					)
					return
				}

				os.Args =
					append(
						os.Args[:i],
						os.Args[i+2:]...,
					)

				break
			}
		}


		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--priority" {
				if i+1 >= len(os.Args) {
					fmt.Println("missing priority")
					return
				}

				switch strings.ToLower(os.Args[i+1]) {
				case "low":
					priority = 1
				case "medium":
					priority = 2
				case "high":
					priority = 3
				default:
					fmt.Println(
						"priority must be low, medium or high",
					)
					return
				}

				os.Args =
					append(
						os.Args[:i],
						os.Args[i+2:]...,
					)

				break
			}
		}

		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--tags" {
				if i+1 >= len(os.Args) {
					fmt.Println("missing tags")
					return
				}

				tags = os.Args[i+1]

				os.Args =
					append(
						os.Args[:i],
						os.Args[i+2:]...,
					)

				break
			}
		}

		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--due" {
				if i+1 >= len(os.Args) {
					fmt.Println("missing date")
					return
				}

				d, err :=
					time.Parse(
						"2006-01-02",
						os.Args[i+1],
					)
				if err != nil {
					fmt.Println(
						"date format: YYYY-MM-DD",
					)
					return
				}

				due = &d

				os.Args =
					append(
						os.Args[:i],
						os.Args[i+2:]...,
					)

				break
			}
		}
		if recurring != "" && due == nil {
			fmt.Println(
				"recurring tasks require --due YYYY-MM-DD",
			)
			return
		}
		
		text := strings.Join(os.Args[2:], " ")

		err := task.AddTask(
			database,
			text,
			priority,
			tags,
			due,
			recurring,
		)
		if err != nil {
			panic(err)
		}

		fmt.Println("Task added.")

	case "list":
		tasks, err := task.List(database)
		if err != nil {
			panic(err)
		}

		printTasks(tasks)
	

	case "today":
		tasks, err :=
			task.Today(database)
		if err != nil {
			panic(err)
		}

		printTasks(tasks)

	case "overdue":
		tasks, err :=
			task.Overdue(database)
		if err != nil {
			panic(err)
		}

		printTasks(tasks)
	
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("usage: td delete <id>")
			return
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("invalid id")
			return
		}

		err = task.Delete(database, id)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("task not found")
				return
			}
			panic(err)
		}

		fmt.Println("Task deleted.")

	case "modify":
		if len(os.Args) < 4 {
			fmt.Println("usage: td modify <id> <task>")
			return
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("invalid id")
			return
		}

		text := strings.Join(os.Args[3:], " ")

		err = task.Modify(database, id, text)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("task not found")
				return
			}
			panic(err)
		}

		fmt.Println("Task updated.")

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("usage: td search <text>")
			return
		}

		query := strings.Join(
			os.Args[2:],
			" ",
		)

		tasks, err :=
			task.Search(
				database,
				query,
			)
		if err != nil {
			panic(err)
		}

		printTasks(tasks)

	case "done":
		if len(os.Args) != 3 {
			fmt.Println("usage: td done <id>")
			return
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("invalid id")
			return
		}

		err = task.Done(database, id)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("task not found")
				return
			}
			panic(err)
		}

		fmt.Println("Task completed.")

	case "undo":
		if len(os.Args) != 3 {
			fmt.Println("usage: td undo <id>")
			return
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("invalid id")
			return
		}

		t, err := task.Get(database, id)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("task not found")
				return
			}
			panic(err)
		}

		if t.Recurring != "" {
			fmt.Println(
				"cannot undo recurring task occurrences",
			)
			return
		}

		err = task.Undo(database, id)
		if err != nil {
			panic(err)
		}

		fmt.Println("Task marked pending.")

	case "path":
		fmt.Println(path)
	
	case "version":
		fmt.Println(Version)		

	default:
		usage()
	}
}