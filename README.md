# td

A lightweight terminal task manager written in Go.

`td` stores your tasks locally using SQLite and follows the Linux XDG directory specification.

## Features

- Simple CLI interface
- SQLite persistence
- Bulk task creation using your preferred editor
- XDG-compliant data and config storage
- Single binary, no server required
- Native Debian package

---

## Installation

### Debian Package

```bash
sudo dpkg -i td_1.0.0_amd64.deb
```

---

## Build from Source

Requirements:

- Go 1.24+

Clone and build:

```bash
git clone https://github.com/<your-username>/td.git
cd td

go build -o td .
```

Install locally:

```bash
sudo install -m755 td /usr/local/bin/td
```

Verify:

```bash
td
```

---

## Usage

### Add a task

```bash
td add "Learn Kubernetes"
```

### List tasks

```bash
td list
```

Example:

```text
=============== Tasks ===============
1. Learn Kubernetes
2. Build a Go CLI
3. Read Distributed Systems book
```

### Modify a task

```bash
td modify 2 "Build and package a Go CLI"
```

### Delete a task

```bash
td delete 1
```

---

# Bulk Add Mode

Running:

```bash
td add
```

opens your preferred editor.

Example:

```text
Learn Go
Learn Kubernetes
Buy milk
Renew SSL certificate
```

Save and exit.

Every non-empty line becomes a task.

---

# Configure Preferred Editor

```bash
td add -e
```

Example:

```text
Editor: micro
Editor updated.
```

Supported editors:

- micro
- vim
- nvim
- nano
- helix
- emacs
- anything available on your system

`td` respects the following order:

1. Configured editor (`~/.config/td/config.json`)
2. `$EDITOR`
3. `nano`
4. `vi`

---

# Data Storage

Tasks database:

```text
~/.local/share/td/tasks.db
```

Configuration:

```text
~/.config/td/config.json
```

This keeps your data separate from the installed binary and follows Linux conventions.

---

# Example Workflow

```bash
td add "Learn Go"
td add "Learn Kubernetes"
td list

td add
# opens editor

td modify 2 "Learn Kubernetes deeply"

td delete 1
```

---

# Backup

Database file:

```bash
cp ~/.local/share/td/tasks.db ~/tasks-backup.db
```

Restore:

```bash
cp ~/tasks-backup.db ~/.local/share/td/tasks.db
```

---

# Remove Package

```bash
sudo apt remove td
```

This removes the binary but keeps:

```text
~/.local/share/td/tasks.db
~/.config/td/config.json
```

To completely remove all data:

```bash
rm -rf ~/.local/share/td
rm -rf ~/.config/td
```

---

# Project Structure

```text
.
├── config
├── db
├── editor
├── task
├── main.go
└── README.md
```

---

# License

MIT

---

Built with Go ❤️