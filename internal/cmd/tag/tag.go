// Package tag provides commands for tag operations.
package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewListCmd creates the list command.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewTagLogic()
			if err != nil {
				return err
			}

			tags, err := l.List(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(tags)
			}

			rows := make([][]string, len(tags))
			for i, tag := range tags {
				rows[i] = []string{tag.Label, fmt.Sprintf("%d", tag.Count)}
			}
			output.AsTable([]string{"LABEL", "COUNT"}, rows)
			return nil
		},
	}
}

// NewSearchCmd creates the search command.
func NewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <keyword>",
		Short: "Search for tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewTagLogic()
			if err != nil {
				return err
			}

			result, err := l.Search(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			output.Printf("Keyword: %s\n", result.K)
			output.Println("Tags:")
			for _, tag := range result.Tags {
				output.Printf("  - %s\n", tag)
			}
			return nil
		},
	}
}

// NewDocsCmd creates the docs command.
func NewDocsCmd() *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Get documents with a specific tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" {
				return fmt.Errorf("--label is required")
			}

			l, err := logic.NewTagLogic()
			if err != nil {
				return err
			}

			docs, err := l.GetDocsByTag(cmd.Context(), label)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(docs)
			}

			for _, doc := range docs {
				output.Println(doc)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "Tag label (required)")
	return cmd
}

// NewRenameCmd creates the rename command.
func NewRenameCmd() *cobra.Command {
	var oldLabel, newLabel string
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if oldLabel == "" || newLabel == "" {
				return fmt.Errorf("both --old and --new are required")
			}

			l, err := logic.NewTagLogic()
			if err != nil {
				return err
			}

			err = l.Rename(cmd.Context(), oldLabel, newLabel)
			if err != nil {
				return err
			}

			output.Printf("Tag renamed from '%s' to '%s'\n", oldLabel, newLabel)
			return nil
		},
	}
	cmd.Flags().StringVar(&oldLabel, "old", "", "Current tag label")
	cmd.Flags().StringVar(&newLabel, "new", "", "New tag label")
	return cmd
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	var label string
	var yesFlag bool
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			if label == "" {
				return fmt.Errorf("--label is required")
			}

			if !yesFlag {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewTagLogic()
			if err != nil {
				return err
			}

			err = l.Remove(cmd.Context(), label)
			if err != nil {
				return err
			}

			output.Printf("Tag '%s' removed successfully\n", label)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "Tag label to remove (required)")
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Confirm removal")
	return cmd
}

// NewTagCmd creates the tag command group.
func NewTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Tag operations",
		Long:  "Commands for managing SiYuan tags.",
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewSearchCmd())
	cmd.AddCommand(NewDocsCmd())
	cmd.AddCommand(NewRenameCmd())
	cmd.AddCommand(NewRemoveCmd())

	return cmd
}
