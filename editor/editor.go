package editor

import (
	"os"
	"os/exec"
	"strings"

	"github.com/RajanCodesDev/td/config"
)

func GetEditor() string {
	cfg, _ := config.Load()

	if cfg.Editor != "" {
		return cfg.Editor
	}

	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}

	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}

	return "vi"
}

func SetEditor(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.Editor = strings.TrimSpace(name)

	return config.Save(cfg)
}