package main

import (
	"fmt"
	"gocli/db"
	"gocli/editor"
	"gocli/task"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
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

func usage() {
	fmt.Println(`
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

func printTasks(tasks []task.Task) {
	fmt.Printf(
		"%-3s %-3s %-3s %-12s %-20s %s\n",
		"ID",
		"S",
		"P",
		"Due",
		"Tags",
		"Task",
	)

	fmt.Println("--------------------------------------------------------------------------------")

	for _, t := range tasks {
		status := "○"

		if t.Completed {
			status = Green + "✓" + Reset
		}

		priority := "M"

		switch t.Priority {
		case 1:
			priority = Cyan + "L" + Reset
		case 2:
			priority = Yellow + "M" + Reset
		case 3:
			priority = Red + "H" + Reset
		}

		due := "-"

		if t.DueDate != nil {
			due =
				t.DueDate.Format(
					"2006-01-02",
				)

			now := time.Now()

			if !t.Completed {
				if t.DueDate.Before(now) &&
					!sameDay(*t.DueDate, now) {

					due =
						Red +
						due +
						Reset

				} else if sameDay(
					*t.DueDate,
					now,
				) {

					due =
						Yellow +
						due +
						Reset
				}
			}
		}

		fmt.Printf(
			"%-3d %-3s %-3s %-12s %-20s %s\n",
			t.ID,
			status,
			priority,
			due,
			t.Tags,
			t.Task,
		)
	}
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

		text := strings.Join(os.Args[2:], " ")

		err := task.AddTask(
			database,
			text,
			priority,
			tags,
			due,
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

		err = task.Undo(database, id)
		if err != nil {
			panic(err)
		}

		fmt.Println("Task marked pending.")

	case "path":
		fmt.Println(path)

	default:
		usage()
	}
}