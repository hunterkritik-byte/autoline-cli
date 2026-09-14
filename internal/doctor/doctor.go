package doctor

import (
	"fmt"
	"os/exec"
)

type Check struct {
	Name    string `json:"name"`
	Found   bool   `json:"found"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

func Run() []Check {
	commands := []struct {
		name string
		cmd  string
		args []string
	}{
		{"Git", "git", []string{"--version"}},
		{"Docker", "docker", []string{"--version"}},
		{"Docker Buildx", "docker", []string{"buildx", "version"}},
		{"Pre-commit", "pre-commit", []string{"--version"}},
	}
	checks := make([]Check, 0, len(commands))
	for _, item := range commands {
		path, err := exec.LookPath(item.cmd)
		if err != nil {
			checks = append(checks, Check{Name: item.name, Found: false, Message: "not installed or not on PATH"})
			continue
		}
		if err := exec.Command(item.cmd, item.args...).Run(); err != nil {
			checks = append(checks, Check{Name: item.name, Found: false, Path: path, Message: fmt.Sprintf("command failed: %v", err)})
			continue
		}
		checks = append(checks, Check{Name: item.name, Found: true, Path: path, Message: "available"})
	}
	return checks
}

func Healthy(checks []Check) bool {
	for _, check := range checks {
		if !check.Found {
			return false
		}
	}
	return true
}
