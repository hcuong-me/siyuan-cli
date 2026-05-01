// Package doc provides commands for document operations.
package doc

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewListCmd creates the list command.
func NewListCmd() *cobra.Command {
	var maxDepth int

	return &cobra.Command{
		Use:   "list <notebook-id>",
		Short: "List documents in a notebook",
		Long:  "List all documents in a notebook as a tree structure. Use notebook ID from 'notebook list'.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewDocumentLogic()
			if err != nil {
				return err
			}

			tree, err := l.ListTree(cmd.Context(), args[0], maxDepth)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(tree)
			}

			printTree(tree, "")
			return nil
		},
	}
}

func printTree(nodes []logic.DocumentInfo, prefix string) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		symbol := "├── "
		if isLast {
			symbol = "└── "
		}
		name := node.Name
		if name == "" {
			name = node.ID
		}
		output.Printf("%s%s%s\n", prefix, symbol, name)

		if len(node.Children) > 0 {
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			printTree(node.Children, newPrefix)
		}
	}
}

// NewGetCmd creates the get command.
func NewGetCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "get <notebook-id>",
		Short: "Get document content as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewDocumentLogic()
			if err != nil {
				return err
			}

			content, err := l.Get(cmd.Context(), args[0], path)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(content)
			}

			output.Println(content.Content)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Document path (required)")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// NewCreateCmd creates the create command.
func NewCreateCmd() *cobra.Command {
	var path, content, contentFile string
	cmd := &cobra.Command{
		Use:   "create <notebook-id>",
		Short: "Create a new document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewDocumentLogic()
			if err != nil {
				return err
			}

			markdown := content
			if contentFile != "" {
				data, err := os.ReadFile(contentFile)
				if err != nil {
					return fmt.Errorf("failed to read content file: %w", err)
				}
				markdown = string(data)
			}

			docID, err := l.Create(cmd.Context(), args[0], path, markdown)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(map[string]string{"id": docID})
			}

			output.Printf("Created document: %s (ID: %s)\n", path, docID)
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Document path (required)")
	cmd.Flags().StringVar(&content, "content", "", "Document content (Markdown)")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Path to file containing content")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// NewUpdateCmd creates the update command.
func NewUpdateCmd() *cobra.Command {
	var path, content, contentFile string
	cmd := &cobra.Command{
		Use:   "update <notebook-id>",
		Short: "Update document content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewDocumentLogic()
			if err != nil {
				return err
			}

			markdown := content
			if contentFile != "" {
				data, err := os.ReadFile(contentFile)
				if err != nil {
					return fmt.Errorf("failed to read content file: %w", err)
				}
				markdown = string(data)
			}

			err = l.Update(cmd.Context(), args[0], path, markdown)
			if err != nil {
				return err
			}

			output.Println("Document updated successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Document path (required)")
	cmd.Flags().StringVar(&content, "content", "", "New content (Markdown)")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Path to file containing content")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	var path string
	var yesFlag bool
	cmd := &cobra.Command{
		Use:   "remove <notebook-id>",
		Short: "Remove a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yesFlag {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewDocumentLogic()
			if err != nil {
				return err
			}

			err = l.Remove(cmd.Context(), args[0], path)
			if err != nil {
				return err
			}

			output.Println("Document removed successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "path", "", "Document path (required)")
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Confirm removal")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// NewDocCmd creates the doc command group.
func NewDocCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doc",
		Short:   "Document operations",
		Aliases: []string{"document"},
		Long:    "Commands for managing SiYuan documents.",
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewRemoveCmd())

	return cmd
}
