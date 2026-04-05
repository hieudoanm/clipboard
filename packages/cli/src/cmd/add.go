package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hieudoanm/clipboard/src/db"
	"github.com/spf13/cobra"
)

var addSource string

var addCmd = &cobra.Command{
	Use:   "add [content]",
	Short: "Add a clip to the clipboard history",
	Long: `Add text to clipboard history. Content can be passed as an argument
or piped via stdin.

Examples:
  clips add "hello world"
  echo "hello world" | clips add
  clips add --source "browser" "some URL"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var content string

		if len(args) > 0 {
			content = strings.Join(args, " ")
		} else {
			// Read from stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("provide content as argument or via stdin")
			}
			var sb strings.Builder
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if sb.Len() > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(scanner.Text())
			}
			content = sb.String()
		}

		path := dbPath
		if path == "" {
			path = db.DefaultPath()
		}
		database, err := db.Open(path)
		if err != nil {
			return err
		}
		defer database.Close()

		entry, err := database.Add(content, addSource)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Saved clip #%d\n", entry.ID)
		preview := entry.Content
		if len(preview) > 80 {
			preview = preview[:77] + "..."
		}
		fmt.Printf("  %s\n", preview)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addSource, "source", "", "source label (e.g. browser, terminal)")
	rootCmd.AddCommand(addCmd)
}
