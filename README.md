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
- Zero configuration
- Single binary

---


# Installation

## Ubuntu (PPA)

Add the official PPA:

```bash
sudo add-apt-repository ppa:rajancodesdev/td
sudo apt update
sudo apt install td
```

Verify the installation:

```bash
td version
```

---

## Build From Source

Requirements:

- Go 1.25+
- GCC (required by SQLite)

```bash
git clone https://github.com/RajanCodesDev/td.git
cd td

go build -o td

sudo mv td /usr/local/bin/
```

Verify:

```bash
td version
```

---


Requirements:

- Go 1.25+
- GCC (required by SQLite)

```bash
git clone https://github.com/RajanCodesDev/td.git

cd td

go build -o td

sudo mv td /usr/local/bin/
```

Verify:

```bash
td version
```

---

## Database

The database is created automatically on first run.

```
~/.local/share/td/tasks.db
```

Configuration is stored in:

```
~/.config/td/config.json
```

---

# Usage

## Add Tasks

```bash
td add "Buy groceries"

td add "Fix production bug" -p work

td add "Review Kubernetes docs" --project learning

td add "Pay rent" --priority high

td add "Renew domain" --due 2026-08-01

td add "Daily journal" --due 2026-08-01 --every daily

td add "Deploy cluster" \
    -p work \
    --priority high \
    --due 2026-08-01
```

Open your preferred editor for bulk task creation:

```bash
td add -e
```

---

## List Tasks

```bash
td list

# alias
td ls
```

---

## Projects

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

## Complete a Task

```bash
td done 3

# alias
td c 3
```

---

## Undo a Task

```bash
td undo 3

# alias
td u 3
```

---

## Modify a Task

```bash
td modify 3 "Updated task"
```

---

## Delete a Task

```bash
td delete 3

# aliases
td rm 3
td del 3
```

---

## Search

Searches both task titles and project names.

```bash
td search kubernetes

# alias
td s kubernetes
```

---

## Due Dates

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

## Statistics

```bash
td stats

# alias
td stat
```

---

## Clear Completed Tasks

```bash
td clear-completed

# alias
td cc
```

---

## Version

```bash
td version

# alias
td v
```

---

# Recurring Tasks

Supported schedules:

```
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

# Priorities

```
High
Medium
Low
```

Default priority is **Medium**.

---

# Task Ordering

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

# Data Locations

Database

```
~/.local/share/td/tasks.db
```

Configuration

```
~/.config/td/config.json
```

---

# License

MIT