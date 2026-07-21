# td

A fast, lightweight terminal task manager written in Go and backed by SQLite.

## Features

* SQLite persistence
* Priorities (Low, Medium, High)
* project
* Due dates
* Today and Overdue views
* Recurring tasks
* Search
* Statistics
* Bulk task creation via editor
* Pretty terminal tables
* Command aliases

---

## Installation

```bash
git clone https://github.com/<your-username>/td.git
cd td
go build -o td
sudo mv td /usr/local/bin/
```

---

## Usage

### Add Tasks

```bash
td add "Buy groceries"
td add "Pay rent" --priority high
td add "Deploy cluster" --project work,kubernetes
td add "Renew domain" --due 2026-08-01
td add "Daily journal" --due 2026-08-01 --every daily
```

### List Tasks

```bash
td list
```

### Complete a Task

```bash
td done 3
```

### Undo a Task

```bash
td undo 3
```

### Modify a Task

```bash
td modify 3 "Updated task"
```

### Delete a Task

```bash
td delete 3
```

### Search

```bash
td search journal
```

### Due Dates

```bash
td today
td overdue
```

### Statistics

```bash
td stats
```

### Clear Completed Tasks

```bash
td clear-completed
```

---

## Recurring Tasks

Supported recurrence schedules:

```bash
--every daily
--every weekly
--every monthly
--every yearly
```

Example:

```bash
td add "Daily backup" --due 2026-08-01 --every daily
```

Completing a recurring task automatically creates the next occurrence.

---

## Aliases

```text
td ls          list tasks
td c <id>      complete task
td u <id>      undo task
td rm <id>     delete task
td t           today
td o           overdue
td s <text>    search
td stat        statistics
td cc          clear completed
td v           version
```

---

## Task Ordering

Tasks are automatically sorted in the following order:

1. Overdue
2. Today
3. Tomorrow
4. Future due dates
5. No due date
6. Completed

Within the same group:

1. Higher priority first
2. Earlier due date first
3. Lower ID first

---

## Database Location

```text
~/.local/share/td/tasks.db
```

---

## License

MIT
