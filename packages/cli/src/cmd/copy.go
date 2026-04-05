package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <id>",
	Short: "Copy a clip to the system clipboard",
	Long: `Copy a stored clip to your OS clipboard so you can paste it anywhere.

Supported platforms: macOS (pbcopy), Linux (xclip/xsel/wl-copy), Windows (clip).

Examples:
  clips copy 3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid id: %s", args[0])
		}
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		e, err := database.Get(id)
		if err != nil {
			return fmt.Errorf("clip %d not found", id)
		}

		if err := copyToClipboard(e.Content); err != nil {
			// Graceful fallback: print content so user can copy manually
			fmt.Printf("⚠ Could not access system clipboard (%v)\n", err)
			fmt.Println("Content:")
			fmt.Println(strings.Repeat("─", 40))
			fmt.Println(e.Content)
			return nil
		}

		preview := e.Content
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		fmt.Printf("✓ Copied clip #%d to clipboard\n  %s\n", e.ID, preview)
		return nil
	},
}

func copyToClipboard(content string) error {
	var copyCmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		copyCmd = exec.Command("pbcopy")
	case "windows":
		copyCmd = exec.Command("clip")
	default: // Linux / BSD
		// Try wl-copy (Wayland) first, then xclip, then xsel
		for _, tool := range []string{"wl-copy", "xclip", "xsel"} {
			if _, err := exec.LookPath(tool); err == nil {
				switch tool {
				case "wl-copy":
					copyCmd = exec.Command("wl-copy")
				case "xclip":
					copyCmd = exec.Command("xclip", "-selection", "clipboard")
				case "xsel":
					copyCmd = exec.Command("xsel", "--clipboard", "--input")
				}
				break
			}
		}
		if copyCmd == nil {
			return fmt.Errorf("no clipboard tool found (install xclip, xsel, or wl-copy)")
		}
	}

	in, err := copyCmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := copyCmd.Start(); err != nil {
		return err
	}
	fmt.Fprint(in, content)
	in.Close()
	return copyCmd.Wait()
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
