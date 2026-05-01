// Package system provides commands for system operations.
package system

import (
	"github.com/spf13/cobra"

	"siyuan/internal/siyuan"
	"siyuan/internal/utils/output"
)

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get SiYuan system version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := siyuan.New()
			if err != nil {
				return err
			}

			version, err := c.GetVersion(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(version)
			}

			output.Println(version)
			return nil
		},
	}
}

// NewTimeCmd creates the time command.
func NewTimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "time",
		Short: "Get SiYuan system current time",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := siyuan.New()
			if err != nil {
				return err
			}

			timestamp, err := c.GetCurrentTime(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(timestamp)
			}

			output.Println(timestamp)
			return nil
		},
	}
}

// NewBootProgressCmd creates the boot-progress command.
func NewBootProgressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "boot-progress",
		Short: "Get SiYuan boot progress",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := siyuan.New()
			if err != nil {
				return err
			}

			progress, err := c.GetBootProgress(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(progress)
			}

			output.Printf("%s (%d%%)\n", progress.Details, progress.Progress)
			return nil
		},
	}
}

// NewSystemCmd creates the system command group.
func NewSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System operations",
		Long:  "Commands for querying SiYuan system status.",
	}

	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewTimeCmd())
	cmd.AddCommand(NewBootProgressCmd())

	return cmd
}
