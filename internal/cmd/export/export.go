// Package export provides commands for export operations.
package export

import (
	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewMarkdownCmd creates the markdown export command.
func NewMarkdownCmd() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "markdown <doc-id>",
		Short: "Export a document as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewExportLogic()
			if err != nil {
				return err
			}

			result, err := l.ExportMarkdown(cmd.Context(), args[0], outputPath)
			if err != nil {
				return err
			}

			if outputPath != "" {
				output.Printf("Exported to %s\n", outputPath)
			} else {
				output.Println(result.Content)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path")
	return cmd
}

// NewHTMLCmd creates the HTML export command.
func NewHTMLCmd() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "html <doc-id>",
		Short: "Export a document as HTML",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewExportLogic()
			if err != nil {
				return err
			}

			result, err := l.ExportHTML(cmd.Context(), args[0], outputPath)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			if outputPath != "" {
				output.Printf("Exported to %s\n", outputPath)
			} else {
				output.Println(result.Content)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path")
	return cmd
}

// NewPDFCmd creates the PDF export command.
func NewPDFCmd() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "pdf <doc-id>",
		Short: "Export a document as PDF (server-side)",
		Long: `Export a document as PDF.

Note: The PDF is generated on the server. The returned path is relative to the server.
You can download the PDF from SiYuan's export directory.

Examples:
  siyuan export pdf 20260316120000-abc123
  siyuan export pdf 20260316120000-abc123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewExportLogic()
			if err != nil {
				return err
			}

			result, err := l.ExportPDF(cmd.Context(), args[0], outputPath)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			output.Printf("PDF exported to server path: %s\n", result.Path)
			output.Println("Note: Download the file from SiYuan's export directory.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (server-side)")
	return cmd
}

// NewDocxCmd creates the DOCX export command.
func NewDocxCmd() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "docx <doc-id>",
		Short: "Export a document as DOCX (server-side)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewExportLogic()
			if err != nil {
				return err
			}

			result, err := l.ExportDocx(cmd.Context(), args[0], outputPath)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			output.Printf("DOCX exported to server path: %s\n", result.Path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (not yet implemented)")
	return cmd
}

// NewPreviewCmd creates the preview command.
func NewPreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview <doc-id>",
		Short: "Preview a document (same as get)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewExportLogic()
			if err != nil {
				return err
			}

			result, err := l.ExportMarkdown(cmd.Context(), args[0], "")
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			output.Println(result.Content)
			return nil
		},
	}
}

// NewExportCmd creates the export command group.
func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export operations",
		Long:  "Commands for exporting SiYuan documents to various formats.",
	}

	cmd.AddCommand(NewPreviewCmd())
	cmd.AddCommand(NewMarkdownCmd())
	cmd.AddCommand(NewHTMLCmd())
	cmd.AddCommand(NewPDFCmd())
	cmd.AddCommand(NewDocxCmd())

	return cmd
}
