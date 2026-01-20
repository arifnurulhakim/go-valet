package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arifnurulhakim/go-valet/internal/types"
)

const (
	configDirName = "go-valet"
	caddyFileName = "Caddyfile"
)

func GenerateCaddyfile(apps []types.App) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, caddyFileName)

	var sb strings.Builder
	sb.WriteString("{\n\tauto_https off\n}\n\n")

	for _, app := range apps {
		sb.WriteString(fmt.Sprintf("http://%s.test {\n", app.Name))
		sb.WriteString(fmt.Sprintf("\treverse_proxy localhost:%d\n", app.Port))
		sb.WriteString("}\n\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func Reload() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(configDir, configDirName, caddyFileName)

	// Check if caddy is running? Or just try reload. If not running, start it?
	// The daemon will be running. But caddy is a separate process.
	// We can use `caddy reload` if it's already running.
	// But first run, we might need `caddy run`.
	// For simplicity, let's assume `caddy start` or `caddy reload` works.
	// Actually `caddy reload` requires the admin API, which I turned off above (`admin off`).
	// Ah, if `admin off` is set, `caddy reload` might not work via CLI if it uses the API.
	// The `caddy reload` command uses the admin API.
	// So I should NOT turn off admin if I want to use `caddy reload`.
	// Or I can leave admin on default (localhost:2019).

	// Let's remove `admin off` to allow reloading.

	cmd := exec.Command("caddy", "reload", "--config", path)
	if err := cmd.Run(); err != nil {
		// If reload fails, maybe it's not running. Try start.
		cmd = exec.Command("caddy", "start", "--config", path)
		return cmd.Run()
	}
	return nil
}
