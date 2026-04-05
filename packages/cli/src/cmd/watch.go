package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var watchInterval int

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitor the OS clipboard and auto-save new copies",
	Long: `Continuously poll the system clipboard and save any new content to history.
Press Ctrl+C to stop.

Examples:
  clips watch                  # poll every 500ms (default)
  clips watch --interval 1000  # poll every 1000ms`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		interval := time.Duration(watchInterval) * time.Millisecond
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Graceful shutdown on Ctrl+C
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		fmt.Printf("👀 Watching clipboard (every %dms) — Ctrl+C to stop\n", watchInterval)
		fmt.Println(strings.Repeat("─", 50))

		var last string

		// Read once immediately before the first tick
		if current, err := readClipboard(); err == nil {
			last = current
		}

		for {
			select {
			case <-sig:
				fmt.Println("\n✓ Stopped watching.")
				return nil

			case <-ticker.C:
				current, err := readClipboard()
				if err != nil {
					// Clipboard may be temporarily unavailable; skip silently
					continue
				}
				current = strings.TrimSpace(current)
				if current == "" || current == last {
					continue
				}
				last = current

				entry, err := database.Add(current, "watch")
				if err != nil {
					fmt.Fprintf(os.Stderr, "⚠ save error: %v\n", err)
					continue
				}

				preview := current
				if len(preview) > 60 {
					preview = preview[:57] + "..."
				}
				// Replace newlines for single-line display
				preview = strings.ReplaceAll(preview, "\n", " ↵ ")

				fmt.Printf("[%s] #%-4d  %s\n",
					time.Now().Format("15:04:05"),
					entry.ID,
					preview,
				)
			}
		}
	},
}

// readClipboard reads the current system clipboard content.
func readClipboard() (string, error) {
	var out []byte
	var err error

	switch runtime.GOOS {
	case "darwin":
		out, err = exec.Command("pbpaste").Output()
	case "windows":
		out, err = exec.Command(
			"powershell", "-noprofile", "-command",
			"Get-Clipboard",
		).Output()
	default: // Linux / BSD
		for _, tool := range []struct{ cmd, arg string }{
			{"wl-paste", "--no-newline"},
			{"xclip", "-selection clipboard -o"},
			{"xsel", "--clipboard --output"},
		} {
			if _, lerr := exec.LookPath(tool.cmd); lerr != nil {
				continue
			}
			parts := strings.Fields(tool.cmd + " " + tool.arg)
			out, err = exec.Command(parts[0], parts[1:]...).Output()
			break
		}
		if out == nil && err == nil {
			return "", fmt.Errorf("no clipboard tool found")
		}
	}

	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func init() {
	watchCmd.Flags().IntVarP(&watchInterval, "interval", "i", 500, "polling interval in milliseconds")

	rootCmd.AddCommand(watchCmd)
}
