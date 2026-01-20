package daemon

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/arifnurulhakim/go-valet/internal/caddy"
	"github.com/arifnurulhakim/go-valet/internal/discovery"
	"github.com/arifnurulhakim/go-valet/internal/registry"
	"github.com/arifnurulhakim/go-valet/internal/runner"
	"github.com/arifnurulhakim/go-valet/internal/types"
)

type Daemon struct {
	processes map[string]*runner.Process // keyed by App Name
	mu        sync.Mutex
	watcher   *fsnotify.Watcher
}

func Start() {
	d := &Daemon{
		processes: make(map[string]*runner.Process),
	}

	// Setup fsnotify
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	d.watcher = w
	defer w.Close()

	// Watch registry file
	configDir, _ := os.UserConfigDir()
	regFile := filepath.Join(configDir, "go-valet", "registry.json")
	if err := w.Add(filepath.Dir(regFile)); err != nil {
		log.Printf("Warning: could not watch registry dir: %v", err)
	}

	// Handle Updates
	updateChan := make(chan bool, 1) // Buffered to coalesce

	// Initial Sync
	d.Sync()

	// Watch loop
	go func() {
		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return
				}
				// If registry changed or something in a park changed
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					// Debounce: verify if we should sync
					// For now, just trigger. Coalescing happening in trigger.
					select {
					case updateChan <- true:
					default:
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Debounce Loop
	go func() {
		for range updateChan {
			time.Sleep(500 * time.Millisecond) // Debounce window
			// Drain others
		drain:
			for {
				select {
				case <-updateChan:
				default:
					break drain
				}
			}
			d.Sync()
		}
	}()

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("Shutting down...")
	d.Shutdown()
}

func (d *Daemon) Sync() {
	d.mu.Lock()
	defer d.mu.Unlock()

	log.Println("Syncing...")

	reg := registry.New()
	if err := reg.Load(); err != nil {
		log.Printf("Error loading registry: %v", err)
		return
	}

	apps, err := discovery.FindApps(reg)
	if err != nil {
		log.Printf("Error discovering apps: %v", err)
		return
	}

	// Update Watcher for Parks
	// We should remove old watches? fsnotify doesn't support RemoveAll easily, but we can Remove.
	// For simplicity, we just Add. Add is idempotent-ish (duplicates ignored by most systems or easy to handle).
	// Ideally we track watched paths.
	for _, park := range reg.Parks {
		d.watcher.Add(park)
	}

	// Diff Apps
	newApps := make(map[string]types.App)
	for _, app := range apps {
		newApps[app.Name] = app
	}

	// Stop removed apps
	for name, proc := range d.processes {
		if _, ok := newApps[name]; !ok {
			log.Printf("Stopping %s...", name)
			if err := proc.Stop(); err != nil {
				log.Printf("Error stopping %s: %v", name, err)
			}
			delete(d.processes, name)
		}
	}

	// Start new apps
	var activeApps []types.App
	for name, app := range newApps {
		activeApps = append(activeApps, app)
		if _, ok := d.processes[name]; !ok {
			log.Printf("Starting %s on port %d...", name, app.Port)
			proc := &runner.Process{App: app}
			if err := proc.Start(); err != nil {
				log.Printf("Error starting %s: %v", name, err)
				continue
			}
			d.processes[name] = proc
		} else {
			// Check if process is still running, if not restart
			if !d.processes[name].IsRunning() {
				log.Printf("Restarting crashed app %s...", name)
				proc := &runner.Process{App: app}
				if err := proc.Start(); err != nil {
					log.Printf("Error restarting %s: %v", name, err)
					delete(d.processes, name)
					continue
				}
				d.processes[name] = proc
			}
		}
	}

	// Update Caddy
	if err := caddy.GenerateCaddyfile(activeApps); err != nil {
		log.Printf("Error generating Caddyfile: %v", err)
	}
	if err := caddy.Reload(); err != nil {
		log.Printf("Error reloading Caddy: %v", err)
	}

	// Save State
	var appStates []registry.AppState
	for _, app := range activeApps {
		status := "stopped"
		if p, ok := d.processes[app.Name]; ok && p.IsRunning() {
			status = "running"
		}
		appStates = append(appStates, registry.AppState{
			Name:   app.Name,
			Path:   app.Path,
			Port:   app.Port,
			Status: status,
		})
	}
	if err := registry.SaveState(&registry.State{Apps: appStates}); err != nil {
		log.Printf("Error saving state: %v", err)
	}
}

func (d *Daemon) Shutdown() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for name, proc := range d.processes {
		log.Printf("Stopping %s...", name)
		proc.Stop()
	}
}
