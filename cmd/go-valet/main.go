package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arifnurulhakim/go-valet/internal/daemon"
	"github.com/arifnurulhakim/go-valet/internal/registry"
	"github.com/arifnurulhakim/go-valet/internal/trust"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "go-valet",
		Short: "A local development orchestrator for Go applications",
	}

	var parkCmd = &cobra.Command{
		Use:   "park",
		Short: "Register the current directory as a park",
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()
			reg := registry.New()
			reg.Load()
			if err := reg.AddPark(cwd); err != nil {
				fmt.Printf("Error parking: %v\n", err)
				return
			}
			reg.Save()
			fmt.Printf("Parked %s\n", cwd)
		},
	}

	var unparkCmd = &cobra.Command{
		Use:   "unpark",
		Short: "Unregister the current directory as a park",
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()
			reg := registry.New()
			reg.Load()
			if err := reg.RemovePark(cwd); err != nil {
				fmt.Printf("Error unparking: %v\n", err)
				return
			}
			reg.Save()
			fmt.Printf("Unparked %s\n", cwd)
		},
	}

	var linkCmd = &cobra.Command{
		Use:   "link [name]",
		Short: "Link the current directory to a domain",
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()
			name := filepath.Base(cwd)
			if len(args) > 0 {
				name = args[0]
			}

			reg := registry.New()
			reg.Load()
			if err := reg.AddLink(name, cwd); err != nil {
				fmt.Printf("Error linking: %v\n", err)
				return
			}
			reg.Save()
			fmt.Printf("Linked %s to %s\n", cwd, name)
		},
	}

	var unlinkCmd = &cobra.Command{
		Use:   "unlink [name]",
		Short: "Unlink a domain",
		Run: func(cmd *cobra.Command, args []string) {
			name := ""
			if len(args) > 0 {
				name = args[0]
			} else {
				cwd, _ := os.Getwd()
				name = filepath.Base(cwd)
			}

			reg := registry.New()
			reg.Load()
			reg.RemoveLink(name)
			reg.Save()
			fmt.Printf("Unlinked %s\n", name)
		},
	}

	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Manage the background daemon",
	}

	var daemonStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the daemon",
		Run: func(cmd *cobra.Command, args []string) {
			daemon.Start()
		},
	}

	daemonCmd.AddCommand(daemonStartCmd)

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show status of parks and apps",
		Run: func(cmd *cobra.Command, args []string) {
			reg := registry.New()
			reg.Load()

			fmt.Println("Parks:")
			if len(reg.Parks) == 0 {
				fmt.Println("  (none)")
			}
			for _, p := range reg.Parks {
				fmt.Printf("  - %s\n", p)
			}
			fmt.Println()

			fmt.Println("Links:")
			if len(reg.Links) == 0 {
				fmt.Println("  (none)")
			}
			for name, path := range reg.Links {
				fmt.Printf("  - %s -> %s\n", name, path)
			}
			fmt.Println()

			state, err := registry.GetState()
			if err != nil {
				fmt.Println("Error reading state (daemon might not be running):", err)
				return
			}

			fmt.Println("Apps:")
			if len(state.Apps) == 0 {
				fmt.Println("  (none)")
			}
			for _, app := range state.Apps {
				fmt.Printf("  - http://%s.test :%d (%s) [%s]\n", app.Name, app.Port, app.Status, app.Path)
			}
		},
	}

	var linksCmd = &cobra.Command{
		Use:   "links",
		Short: "List all registered links",
		Run: func(cmd *cobra.Command, args []string) {
			reg := registry.New()
			reg.Load()

			if len(reg.Links) == 0 {
				fmt.Println("No links registered.")
				return
			}

			// Table header
			fmt.Printf("%-20s %-30s %s\n", "Site", "URL", "Path")
			fmt.Printf("%-20s %-30s %s\n", "----", "---", "----")
			for name, path := range reg.Links {
				url := fmt.Sprintf("http://%s.test:9090", name)
				fmt.Printf("%-20s %-30s %s\n", name, url, path)
			}
		},
	}

	var trustCmd = &cobra.Command{
		Use:   "trust",
		Short: "Enable port forwarding (80 -> 9090) to allow clean URLs",
		Run: func(cmd *cobra.Command, args []string) {
			if err := trust.Trust(); err != nil {
				fmt.Printf("Error enabling trust: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var untrustCmd = &cobra.Command{
		Use:   "untrust",
		Short: "Disable port forwarding",
		Run: func(cmd *cobra.Command, args []string) {
			if err := trust.Untrust(); err != nil {
				fmt.Printf("Error disabling trust: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart the daemon",
		Run: func(cmd *cobra.Command, args []string) {
			// Stop
			// We can't easily stop the daemon from here without killing it or sending a signal if we don't have pid.
			// But since we use 'brew services', maybe we should just say "use brew services restart"?
			// Or we can try to find the process?
			// Ideally `daemonStartCmd` should handle "already running" or we act as a client sending signal?
			// For now, let's keep it simple: instruct user or try `daemon.Restart`?
			// Let's implement a simple "killall air & go-valet" logic? No that's abrupt.
			// Let's just invoke brew services if on mac?
			// Or just tell the user.
			// But the user *tried* `go-valet restart` and got error.
			// Let's implement a wrapper.

			fmt.Println("Restarting (via brew services)...")
			exec.Command("brew", "services", "restart", "go-valet").Run()
		},
	}

	rootCmd.AddCommand(parkCmd, unparkCmd, linkCmd, unlinkCmd, linksCmd, daemonCmd, statusCmd, trustCmd, untrustCmd, restartCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
