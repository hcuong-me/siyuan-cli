// Package cmd provides the root command for the siyuan CLI.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"siyuan/internal/cmd/asset"
	"siyuan/internal/cmd/attr"
	"siyuan/internal/cmd/block"
	"siyuan/internal/cmd/doc"
	"siyuan/internal/cmd/export"
	siyfile "siyuan/internal/cmd/file"
	"siyuan/internal/cmd/notebook"
	"siyuan/internal/cmd/search"
	"siyuan/internal/cmd/snapshot"
	sqlcmd "siyuan/internal/cmd/sql"
	"siyuan/internal/cmd/system"
	"siyuan/internal/cmd/tag"
	"siyuan/internal/cmd/template"
	"siyuan/internal/utils/output"
)

var (
	// jsonOutput enables JSON output format
	jsonOutput bool
	// version is the CLI version
	version = "dev"
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "siyuan-cli",
		Short:         "CLI for SiYuan Note",
		SilenceErrors: true, // We handle errors ourselves in Execute()
		SilenceUsage:  true, // Don't show usage on error
		Long: `siyuan-cli is a CLI tool for SiYuan Note.

Search notes, read documents, update content, and export results from the terminal.

Before running commands, set SIYUAN_BASE_URL and SIYUAN_TOKEN environment variables:
  export SIYUAN_BASE_URL="http://127.0.0.1:6806"
  export SIYUAN_TOKEN="your-token"

Get your token from Settings > About in SiYuan.`,
		Version: version,
	}

	// Global flags
	root.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")

	// Add subcommands
	root.AddCommand(system.NewSystemCmd())
	root.AddCommand(notebook.NewNotebookCmd())
	root.AddCommand(doc.NewDocCmd())
	root.AddCommand(block.NewBlockCmd())
	root.AddCommand(tag.NewTagCmd())
	root.AddCommand(attr.NewAttrCmd())
	root.AddCommand(export.NewExportCmd())
	root.AddCommand(search.NewSearchCmd())
	root.AddCommand(sqlcmd.NewSQLCmd())
	root.AddCommand(template.NewTemplateCmd())
	root.AddCommand(snapshot.NewSnapshotCmd())
	root.AddCommand(asset.NewAssetCmd())
	root.AddCommand(siyfile.NewFileCmd())

	return root
}

// Execute runs the root command.
func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		output.Error(err)
		os.Exit(1)
	}
}
