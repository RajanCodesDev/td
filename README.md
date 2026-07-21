# td

A fast, lightweight terminal task manager written in Go and backed by SQLite.

## Features

- SQLite persistence
- Projects
- Priorities (Low, Medium, High)
- Due dates
- Today & Overdue views
- Recurring tasks
- Search
- Statistics
- Bulk task creation via your preferred editor
- Pretty terminal tables
- Command aliases

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

td add "Fix production bug" -p work

td add "Review Kubernetes docs" --project learning

td add "Pay rent" --priority high

td add "Renew domain" --due 2026-08-01

td add "Daily journal" --due 2026-08-01 --every daily

td add "Deploy cluster" -p work --priority high --due 2026-08-01
```

---

### List Tasks

```bash
td list

# alias
td ls
```

---

### Projects

List all projects:

```bash
td -p
```

Example:

```text
Projects

learning (5)
personal (2)
work (14)
```

Show tasks belonging to a project:

```bash
td -p work
```

---

### Complete a Task

```bash
td done 3

# alias
td c 3
```

---

### Undo a Task

```bash
td undo 3

# alias
td u 3
```

---

### Modify a Task

```bash
td modify 3 "Updated task"
```

---

### Delete a Task

```bash
td delete 3

# aliases
td rm 3
td del 3
```

---

### Search

Searches task titles **and project names**.

```bash
td search kubernetes

# alias
td s kubernetes
```

---

### Due Dates

```bash
td today

td overdue
```

Aliases:

```bash
td t
td o
```

---

### Statistics

```bash
td stats

# alias
td stat
```

---

### Clear Completed Tasks

```bash
td clear-completed

# alias
td cc
```

---

### Version

```bash
td version

# alias
td v
```

---

## Recurring Tasks

Supported schedules:

```text
daily
weekly
monthly
yearly
```

Example:

```bash
td add "Daily backup" \
    --due 2026-08-01 \
    --every daily
```

Completing a recurring task automatically creates the next occurrence.

---

## Priorities

```text
Low
Medium
High
```

Default priority is **Medium**.

---

## Task Ordering

Tasks are automatically sorted by:

1. Overdue
2. Today
3. Tomorrow
4. Future due dates
5. No due date
6. Completed

Within each group:

1. Higher priority
2. Earlier due date
3. Lower task ID

---

## Database Location

```text
~/.local/share/td/tasks.db
```

---

## License

MIT