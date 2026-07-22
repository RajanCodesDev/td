package editor

import (
	"os"
	"os/exec"
	"strings"
)

func Open() ([]string, error) {
	editor := GetEditor()

	if editor == "" {
		var err error

		editor, err = ChooseEditor()
		if err != nil {
			return nil, err
		}
	}

	tmp, err := os.CreateTemp("", "td-*.txt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		return nil, err
	}

	var tasks []string

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			tasks = append(tasks, line)
		}
	}

	return tasks, nil
}
