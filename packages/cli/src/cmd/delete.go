package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ── delete ───────────────────────────────────────────────────────────────────

var deleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Delete a clip by ID",
	Aliases: []string{"rm", "remove"},
	Args:    cobra.ExactArgs(1),
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

		if err := database.Delete(id); err != nil {
			return err
		}
		fmt.Printf("✓ Deleted clip #%d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
