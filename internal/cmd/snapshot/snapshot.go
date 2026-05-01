// Package snapshot provides commands for repository snapshot operations.
package snapshot

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// formatSize formats bytes to human-readable size.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTime formats Unix timestamp to readable time.
func formatTime(unix int64) string {
	if unix == 0 {
		return "N/A"
	}
	return time.Unix(unix, 0).Format("2006-01-02 15:04:05")
}

// NewListCmd creates the list snapshots command.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all snapshots",
		Long:  "List all repository snapshots.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSnapshotLogic()
			if err != nil {
				return err
			}

			snapshots, err := l.List(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(snapshots)
			}

			if len(snapshots) == 0 {
				output.Println("No snapshots found.")
				return nil
			}

			output.Printf("Found %d snapshots\n\n", len(snapshots))

			rows := make([][]string, len(snapshots))
			for i, s := range snapshots {
				memo := s.Memo
				if memo == "" {
					memo = "(no memo)"
				}
				rows[i] = []string{s.ID, memo, formatTime(s.Created), formatSize(s.Size)}
			}
			output.AsTable([]string{"ID", "MEMO", "CREATED", "SIZE"}, rows)
			return nil
		},
	}
}

// NewCurrentCmd creates the current snapshot command.
func NewCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current snapshot",
		Long:  "Show information about the current repository snapshot.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSnapshotLogic()
			if err != nil {
				return err
			}

			snapshot, err := l.Current(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(snapshot)
			}

			memo := snapshot.Memo
			if memo == "" {
				memo = "(no memo)"
			}

			output.Printf("ID:      %s\n", snapshot.ID)
			output.Printf("Memo:    %s\n", memo)
			output.Printf("Created: %s\n", formatTime(snapshot.Created))
			output.Printf("Files:   %d\n", snapshot.Count)
			output.Printf("Size:    %s\n", formatSize(snapshot.Size))
			return nil
		},
	}
}

// NewCreateCmd creates the create snapshot command.
func NewCreateCmd() *cobra.Command {
	var memo string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new snapshot",
		Long: `Create a new repository snapshot.

Examples:
  siyuan snapshot create
  siyuan snapshot create --memo "before major update"
  siyuan snapshot create --memo "backup-$(date +%Y%m%d)"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSnapshotLogic()
			if err != nil {
				return err
			}

			snapshot, err := l.Create(cmd.Context(), memo)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(snapshot)
			}

			output.Printf("Snapshot created successfully:\n")
			output.Printf("  ID:      %s\n", snapshot.ID)
			if snapshot.Memo != "" {
				output.Printf("  Memo:    %s\n", snapshot.Memo)
			}
			output.Printf("  Created: %s\n", formatTime(snapshot.Created))
			return nil
		},
	}
	cmd.Flags().StringVar(&memo, "memo", "", "Memo/description for the snapshot")
	return cmd
}

// NewRestoreCmd creates the restore snapshot command.
func NewRestoreCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "restore <id>",
		Short: "Restore to a snapshot",
		Long: `Restore the repository to a specific snapshot.

WARNING: This will revert all data to the snapshot state.
Make sure you have a current backup before proceeding.

Examples:
  siyuan snapshot restore 20260316120000-abc123 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("use --yes to confirm restore (this is a destructive operation)")
			}

			l, err := logic.NewSnapshotLogic()
			if err != nil {
				return err
			}

			if err := l.Restore(cmd.Context(), args[0]); err != nil {
				return err
			}

			output.Println("Repository restored successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm restore")
	return cmd
}

// NewRemoveCmd creates the remove snapshot command.
func NewRemoveCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a snapshot",
		Long: `Remove a repository snapshot.

Examples:
  siyuan snapshot remove 20260316120000-abc123 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewSnapshotLogic()
			if err != nil {
				return err
			}

			if err := l.Remove(cmd.Context(), args[0]); err != nil {
				return err
			}

			output.Println("Snapshot removed successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
	return cmd
}

// NewSnapshotCmd creates the snapshot command group.
func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Snapshot operations",
		Long:  "Commands for managing SiYuan repository snapshots (data backups).",
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewCurrentCmd())
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewRestoreCmd())
	cmd.AddCommand(NewRemoveCmd())

	return cmd
}
