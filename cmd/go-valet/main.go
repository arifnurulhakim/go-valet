package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arifnurulhakim/go-valet/internal/daemon"
	"github.com/arifnurulhakim/go-valet/internal/registry"
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

	rootCmd.AddCommand(parkCmd, unparkCmd, linkCmd, unlinkCmd, daemonCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
