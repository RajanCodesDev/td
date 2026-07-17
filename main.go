package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gocli/db"
	"gocli/task"
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
td list
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
		if len(os.Args) < 3 {
			fmt.Println("task text missing")
			return
		}

		text := strings.Join(os.Args[2:], " ")

		err := task.Add(database, text)
		if err != nil {
			panic(err)
		}

		fmt.Println("Task added.")

	case "list":
		tasks, err := task.List(database)
		if err != nil {
			panic(err)
		}

		fmt.Println("=============== Tasks ===============")

		for _, t := range tasks {
			fmt.Printf("%d. %s\n", t.ID, t.Task)
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

	case "path":
		fmt.Println(path)

	default:
		usage()
	}
}