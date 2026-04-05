package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// ── stats ─────────────────────────────────────────────────────────────────────

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show clipboard statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		total, pinned, err := database.Stats()
		if err != nil {
			return err
		}

		fmt.Println("📋 Clipboard Stats")
		fmt.Println(strings.Repeat("─", 30))
		fmt.Printf("  Total clips:   %d\n", total)
		fmt.Printf("  Pinned:        %d\n", pinned)
		fmt.Printf("  Unpinned:      %d\n", total-pinned)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
