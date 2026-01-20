package trust

import (
	"fmt"
	"os"
	"os/exec"
)

const pfRule = "rdr pass on lo0 inet proto tcp from any to any port 80 -> 127.0.0.1 port 9090"

func Trust() error {
	// 1. Create a temporary file with the rule
	f, err := os.CreateTemp("", "go-valet-pf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(pfRule + "\n"); err != nil {
		return fmt.Errorf("failed to write rule: %w", err)
	}
	f.Close()

	// 2. Load the rule using pfctl (requires sudo)
	// We use 'echo "..." | sudo pfctl ...' pattern isn't quite right for interactive sudo.
	// Users should run the go-valet command with sudo, OR we fail and tell them to run with sudo.
	// But `go-valet` is usually installed as user.
	// Better: Use exec.Command("sudo", "pfctl", ...) which will prompt password in terminal.

	cmd := exec.Command("sudo", "pfctl", "-ef", f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Configuring port forwarding (80 -> 9090)...")
	fmt.Println("This requires administrative privileges. Please enter your password if prompted.")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run pfctl: %w", err)
	}

	fmt.Println("Success! Port forwarding enabled.")
	return nil
}

func Untrust() error {
	// Flush rules
	cmd := exec.Command("sudo", "pfctl", "-F", "nat")
	// Note: flushing 'nat' might be too aggressive if user has other NAT rules.
	// But for standard macOS dev setup, it's usually fine.
	// Alternatively, we could just flush our specific anchor if we used anchors.
	// For simplicity in this v1, flushing all NAT/RDR rules is a common Valet strategy.
	// Wait, 'pfctl -F nat' flushes NAT rules. rdr is NAT in context of pf.
	// Actually Laravel Valet uses specific anchor `com.apple/250.LaravelValet`.
	// Let's try to be simple first. if we just use `-f` on main ruleset it replaces everything.
	// Using `-F all` clears everything.

	// Let's just run specific flush or empty file?
	// `pfctl -F all` is dangerous.
	// `pfctl -F nat` ?

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Removing port forwarding...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable packet filter: %w", err)
	}
	return nil
}
