package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const (
	configDirName    = "go-valet"
	registryFileName = "registry.json"
	stateFileName    = "state.json"
)

type Registry struct {
	Parks []string          `json:"parks"`
	Links map[string]string `json:"links"` // domain -> path
	mu    sync.RWMutex
}

type State struct {
	Apps []AppState `json:"apps"`
}

type AppState struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Port   int    `json:"port"`
	Status string `json:"status"` // "running", "stopped"
}

func New() *Registry {
	return &Registry{
		Parks: []string{},
		Links: make(map[string]string),
	}
}

func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, configDirName, registryFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // New registry
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r)
}

func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, registryFileName)
	return os.WriteFile(path, data, 0644)
}

func (r *Registry) AddPark(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for _, p := range r.Parks {
		if p == absPath {
			return nil // Already exists
		}
	}

	r.Parks = append(r.Parks, absPath)
	return nil
}

func (r *Registry) RemovePark(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	for i, p := range r.Parks {
		if p == absPath {
			r.Parks = append(r.Parks[:i], r.Parks[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *Registry) AddLink(name, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if r.Links == nil {
		r.Links = make(map[string]string)
	}
	r.Links[name] = absPath
	return nil
}

func (r *Registry) RemoveLink(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Links, name)
}

func GetState() (*State, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(configDir, configDirName, stateFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &State{Apps: []AppState{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveState(state *State) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, configDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, stateFileName)
	return os.WriteFile(path, data, 0644)
}
