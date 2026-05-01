// Package search provides commands for search operations.
package search

import (
	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewBlockCmd creates the block search command.
func NewBlockCmd() *cobra.Command {
	var page, size int
	cmd := &cobra.Command{
		Use:   "block <keyword>",
		Short: "Full-text search blocks",
		Long:  "Search for blocks containing the given keyword. Supports pagination.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSearchLogic()
			if err != nil {
				return err
			}

			result, err := l.SearchBlocks(cmd.Context(), args[0], page, size)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			output.Printf("Found %d blocks (page %d of %d)\n", result.MatchedBlockCount, page, result.PageCount)
			output.Println()

			if len(result.Blocks) == 0 {
				output.Println("No results found.")
				return nil
			}

			rows := make([][]string, len(result.Blocks))
			for i, block := range result.Blocks {
				// Truncate content for display
				content := block.Content
				if len(content) > 50 {
					content = content[:47] + "..."
				}
				rows[i] = []string{block.ID, block.Type, content}
			}
			output.AsTable([]string{"ID", "TYPE", "CONTENT"}, rows)
			return nil
		},
	}
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().IntVar(&size, "size", 20, "Results per page")
	return cmd
}

// NewDocCmd creates the doc search command.
func NewDocCmd() *cobra.Command {
	var notebook string
	cmd := &cobra.Command{
		Use:   "doc <keyword>",
		Short: "Search documents",
		Long:  "Search for documents by keyword in their path or title.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSearchLogic()
			if err != nil {
				return err
			}

			results, err := l.SearchDocs(cmd.Context(), args[0], notebook, "")
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(results)
			}

			output.Printf("Found %d documents\n", len(results))
			output.Println()

			if len(results) == 0 {
				output.Println("No results found.")
				return nil
			}

			rows := make([][]string, len(results))
			for i, doc := range results {
				rows[i] = []string{doc.HPath, doc.Box}
			}
			output.AsTable([]string{"PATH", "NOTEBOOK"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&notebook, "notebook", "", "Limit search to specific notebook ID")
	return cmd
}

// NewSearchCmd creates the search command group.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search operations",
		Long:  "Commands for searching SiYuan content.",
	}

	cmd.AddCommand(NewBlockCmd())
	cmd.AddCommand(NewDocCmd())

	return cmd
}
