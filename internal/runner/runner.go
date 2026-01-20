package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/arifnurulhakim/go-valet/internal/types"
)

type Process struct {
	App types.App
	cmd *exec.Cmd
}

func (p *Process) Start() error {
	// Ensure log directory exists
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".config", "go-valet", "logs")
	os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(filepath.Join(logDir, p.App.Name+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Try to find air in common locations
	airPath := "air"
	if _, err := os.Stat("/opt/homebrew/bin/air"); err == nil {
		airPath = "/opt/homebrew/bin/air"
	} else if _, err := os.Stat("/usr/local/bin/air"); err == nil {
		airPath = "/usr/local/bin/air"
	}
	// Also check GOPATH/bin
	if _, err := os.Stat(filepath.Join(home, "go/bin/air")); err == nil {
		airPath = filepath.Join(home, "go/bin/air")
	}

	p.cmd = exec.Command(airPath)
	p.cmd.Dir = p.App.Path
	// Inherit environment + PORT + Ensure PATH has common dirs
	env := os.Environ()
	newPath := fmt.Sprintf("PATH=%s:/opt/homebrew/bin:/usr/local/bin:%s/go/bin", os.Getenv("PATH"), home)
	env = append(env, newPath)
	p.cmd.Env = append(env, fmt.Sprintf("PORT=%d", p.App.Port))
	p.cmd.Stdout = logFile
	p.cmd.Stderr = logFile
	p.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Set process group so we can kill subtree

	return p.cmd.Start()
}

func (p *Process) Stop() error {
	if p.cmd != nil && p.cmd.Process != nil {
		// Kill the process group to kill air and the child app
		return syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
	}
	return nil
}

func (p *Process) IsRunning() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	// Check if process is still alive (signal 0)
	if err := p.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

func (p *Process) Wait() error {
	if p.cmd != nil {
		return p.cmd.Wait()
	}
	return nil
}
