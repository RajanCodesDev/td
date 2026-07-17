package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"gocli/editor"
	"gocli/db"
	"gocli/task"
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
td add "task" --priority medium
td add "task" --priority low
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

		text := strings.Join(os.Args[2:], " ")

		err := task.AddWithPriority(
			database,
			text,
			priority,
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

		fmt.Printf(
			"%-3s %-3s %-3s %s\n",
			"ID",
			"S",
			"P",
			"Task",
		)

		fmt.Println("--------------------------------")

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

			fmt.Printf(
				"%-3d %-3s %-3s %s\n",
				t.ID,
				status,
				priority,
				t.Task,
			)
		}

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