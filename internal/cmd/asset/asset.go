// Package asset provides commands for asset operations.
package asset

import (
	"fmt"
	"os"
	"path/filepath"

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

// NewUploadCmd creates the upload asset command.
func NewUploadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload an asset file",
		Long: `Upload a file (image, attachment, etc.) to SiYuan assets.

The file will be stored in the assets directory and can be referenced in notes.
Returns the asset path that can be used in markdown links.

Examples:
  siyuan asset upload ./image.png
  siyuan asset upload ./document.pdf --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check file exists
			if _, err := os.Stat(filePath); err != nil {
				return fmt.Errorf("file not found: %s", filePath)
			}

			l, err := logic.NewAssetLogic()
			if err != nil {
				return err
			}

			result, err := l.Upload(cmd.Context(), filePath)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			if len(result.SuccMap) > 0 {
				output.Printf("Upload successful:\n")
				for original, saved := range result.SuccMap {
					output.Printf("  %s -> %s\n", original, saved)
				}
			}

			if len(result.ErrFiles) > 0 {
				output.Printf("\nFailed files:\n")
				for _, f := range result.ErrFiles {
					output.Printf("  - %s\n", f)
				}
			}

			return nil
		},
	}
}

// NewUnusedCmd creates the unused assets command.
func NewUnusedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unused",
		Short: "List unused assets",
		Long: `List all asset files that are not referenced by any document.

These files can be safely removed to free up disk space.
Use "siyuan asset clean" to remove all unused assets at once.

Examples:
  siyuan asset unused
  siyuan asset unused --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewAssetLogic()
			if err != nil {
				return err
			}

			assets, err := l.Unused(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(assets)
			}

			if len(assets) == 0 {
				output.Println("No unused assets found.")
				return nil
			}

			var totalSize int64
			for _, a := range assets {
				totalSize += a.Size
			}

			output.Printf("Found %d unused assets (%s total)\n\n", len(assets), formatSize(totalSize))

			rows := make([][]string, len(assets))
			for i, a := range assets {
				name := filepath.Base(a.Path)
				rows[i] = []string{name, formatSize(a.Size)}
			}
			output.AsTable([]string{"NAME", "SIZE"}, rows)
			return nil
		},
	}
}

// NewCleanCmd creates the clean assets command.
func NewCleanCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove all unused assets",
		Long: `Remove all asset files that are not referenced by any document.

WARNING: This operation cannot be undone. The files will be permanently deleted.

Examples:
  siyuan asset clean --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("use --yes to confirm cleanup (this will permanently delete unused assets)")
			}

			l, err := logic.NewAssetLogic()
			if err != nil {
				return err
			}

			if err := l.Clean(cmd.Context()); err != nil {
				return err
			}

			output.Println("Unused assets cleaned successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm cleanup")
	return cmd
}

// NewAssetCmd creates the asset command group.
func NewAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset",
		Short: "Asset operations",
		Long:  "Commands for managing SiYuan assets (files, images, attachments).",
	}

	cmd.AddCommand(NewUploadCmd())
	cmd.AddCommand(NewUnusedCmd())
	cmd.AddCommand(NewCleanCmd())

	return cmd
}
