package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hieudoanm/clipboard/src/db"
	"github.com/spf13/cobra"
)

var (
	listLimit  int
	listSearch string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List clipboard history",
	Long: `Display recent clipboard entries, newest first. Pinned clips always appear at the top.

Examples:
  clips list
  clips list --limit 50
  clips list --search "golang"`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := dbPath
		if path == "" {
			path = db.DefaultPath()
		}
		database, err := db.Open(path)
		if err != nil {
			return err
		}
		defer database.Close()

		entries, err := database.List(listLimit, listSearch)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			if listSearch != "" {
				fmt.Printf("No clips found matching %q\n", listSearch)
			} else {
				fmt.Println("No clips yet. Use `clips add <text>` to add one.")
			}
			return nil
		}

		// Header
		fmt.Printf("%-6s  %-4s  %-20s  %-10s  %s\n", "ID", "PIN", "DATE", "SOURCE", "CONTENT")
		fmt.Println(strings.Repeat("─", 90))

		for _, e := range entries {
			pin := "  "
			if e.Pinned {
				pin = "📌"
			}

			// Format date relative if recent
			dateStr := formatRelative(e.CreatedAt)

			source := e.Source
			if source == "" {
				source = "—"
			}
			if len(source) > 10 {
				source = source[:9] + "…"
			}

			// Truncate content for display
			content := strings.ReplaceAll(e.Content, "\n", " ↵ ")
			if len(content) > 50 {
				content = content[:47] + "..."
			}

			fmt.Printf("%-6d  %-4s  %-20s  %-10s  %s\n",
				e.ID, pin, dateStr, source, content)
		}

		fmt.Printf("\n%d clip(s) shown\n", len(entries))
		return nil
	},
}

func formatRelative(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return t.Format("2006-01-02 15:04")
	}
}

func init() {
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "maximum number of clips to show")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "filter clips by content")
	rootCmd.AddCommand(listCmd)
}
