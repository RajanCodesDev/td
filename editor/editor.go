package editor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/RajanCodesDev/td/config"
)

func GetEditor() string {
	cfg, _ := config.Load()

	if strings.TrimSpace(cfg.Editor) != "" {
		return cfg.Editor
	}

	if e := strings.TrimSpace(os.Getenv("VISUAL")); e != "" {
		return e
	}

	if e := strings.TrimSpace(os.Getenv("EDITOR")); e != "" {
		return e
	}

	return ""
}

func SetEditor(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.Editor = strings.TrimSpace(name)

	return config.Save(cfg)
}

func DetectEditors() []string {
	candidates := []string{
		"code",
		"zed",
		"nvim",
		"vim",
		"nano",
		"micro",
		"hx",
		"vi",
	}

	var editors []string

	for _, editor := range candidates {
		if _, err := exec.LookPath(editor); err == nil {
			editors = append(editors, editor)
		}
	}

	return editors
}

func ChooseEditor() (string, error) {
	editors := DetectEditors()

	if len(editors) == 0 {
		return "", fmt.Errorf("no supported editor found")
	}

	fmt.Println("No editor configured.")
	fmt.Println()

	fmt.Println("Available editors:")
	fmt.Println()

	for i, editor := range editors {
		fmt.Printf("%d. %s\n", i+1, editor)
	}

	fmt.Print("\nChoose editor: ")

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid selection")
	}

	if choice < 1 || choice > len(editors) {
		return "", fmt.Errorf("invalid selection")
	}

	selected := editors[choice-1]

	if err := SetEditor(selected); err != nil {
		return "", err
	}

	fmt.Printf("\n✓ Editor set to %s\n\n", selected)

	return selected, nil
}
