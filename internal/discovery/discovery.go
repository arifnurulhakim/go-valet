package discovery

import (
	"os"
	"path/filepath"

	"github.com/hubton/go-valet/internal/allocator"
	"github.com/hubton/go-valet/internal/registry"
	"github.com/hubton/go-valet/internal/types"
)

func FindApps(reg *registry.Registry) ([]types.App, error) {
	var apps []types.App
	seen := make(map[string]bool)

	// 1. Process Links
	for name, path := range reg.Links {
		if !seen[path] && isGoApp(path) {
			apps = append(apps, types.App{
				Name: name,
				Path: path,
				Port: allocator.GetPort(path),
			})
			seen[path] = true
		}
	}

	// 2. Process Parks
	for _, park := range reg.Parks {
		entries, err := os.ReadDir(park)
		if err != nil {
			continue // Skip invalid parks
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(park, entry.Name())
			if seen[fullPath] {
				continue
			}

			if isGoApp(fullPath) {
				apps = append(apps, types.App{
					Name: entry.Name(),
					Path: fullPath,
					Port: allocator.GetPort(fullPath),
				})
				seen[fullPath] = true
			}
		}
	}

	return apps, nil
}

func isGoApp(path string) bool {
	// Check for main.go
	if _, err := os.Stat(filepath.Join(path, "main.go")); err == nil {
		return true
	}

	// Check for cmd/*/main.go
	cmdPath := filepath.Join(path, "cmd")
	entries, err := os.ReadDir(cmdPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				if _, err := os.Stat(filepath.Join(cmdPath, entry.Name(), "main.go")); err == nil {
					return true
				}
			}
		}
	}

	return false
}
