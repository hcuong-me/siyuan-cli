// Package template provides commands for template operations.
package template

import (
	"fmt"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewListCmd creates the list templates command.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all templates",
		Long:  "List all available template files in the templates directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewTemplateLogic()
			if err != nil {
				return err
			}

			templates, err := l.List(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(templates)
			}

			if len(templates) == 0 {
				output.Println("No templates found.")
				return nil
			}

			output.Printf("Found %d templates\n\n", len(templates))

			rows := make([][]string, len(templates))
			for i, t := range templates {
				rows[i] = []string{t.Name}
			}
			output.AsTable([]string{"NAME"}, rows)
			return nil
		},
	}
}

// NewGetCmd creates the get template command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <path>",
		Short: "Get template content",
		Long: `Get the content of a template file.

The path can be:
  - Just the filename (e.g., "template.md")
  - Full path under data/templates (e.g., "/data/templates/template.md")

Examples:
  siyuan template get template.md
  siyuan template get /data/templates/template.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewTemplateLogic()
			if err != nil {
				return err
			}

			content, err := l.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			// For templates, output raw content (could be piped to file)
			output.Println(content)
			return nil
		},
	}
}

// NewRenderCmd creates the render template command.
func NewRenderCmd() *cobra.Command {
	var docID string
	cmd := &cobra.Command{
		Use:   "render <path>",
		Short: "Render a template",
		Long: `Render a template file to a document.

The path can be:
  - Just the filename (e.g., "template.md")
  - Full path under data/templates (e.g., "/data/templates/template.md")

Examples:
  siyuan template render template.md --id 20260316120000-abc123
  siyuan template render daily-note.md --id 20260316120000-abc123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if docID == "" {
				return fmt.Errorf("--id is required: specify the target document ID")
			}

			l, err := logic.NewTemplateLogic()
			if err != nil {
				return err
			}

			result, err := l.Render(cmd.Context(), docID, args[0])
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
	cmd.Flags().StringVar(&docID, "id", "", "Target document ID (required)")
	return cmd
}

// NewRemoveCmd creates the remove template command.
func NewRemoveCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a template",
		Long: `Remove a template file.

The path can be:
  - Just the filename (e.g., "template.md")
  - Full path under data/templates (e.g., "/data/templates/template.md")

Examples:
  siyuan template remove old-template.md --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewTemplateLogic()
			if err != nil {
				return err
			}

			if err := l.Remove(cmd.Context(), args[0]); err != nil {
				return err
			}

			output.Println("Template removed successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
	return cmd
}

// NewTemplateCmd creates the template command group.
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Template operations",
		Long:  "Commands for managing SiYuan templates.",
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewRenderCmd())
	cmd.AddCommand(NewRemoveCmd())

	return cmd
}
