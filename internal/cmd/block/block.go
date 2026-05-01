// Package block provides commands for block operations.
package block

import (
	"fmt"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewGetCmd creates the get command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <block-id>",
		Short: "Get a block's kramdown source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			kramdown, err := l.GetKramdown(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(map[string]string{
					"id":       args[0],
					"kramdown": kramdown,
				})
			}

			output.Println(kramdown)
			return nil
		},
	}
}

// NewChildrenCmd creates the children command.
func NewChildrenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "children <block-id>",
		Short: "Get child blocks of a parent block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			children, err := l.GetChildren(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(children)
			}

			rows := make([][]string, len(children))
			for i, child := range children {
				rows[i] = []string{child.ID, child.Type, child.SubType}
			}
			output.AsTable([]string{"ID", "TYPE", "SUBTYPE"}, rows)
			return nil
		},
	}
}

// NewUpdateCmd creates the update command.
func NewUpdateCmd() *cobra.Command {
	var content string
	cmd := &cobra.Command{
		Use:   "update <block-id>",
		Short: "Update a block's content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			err = l.Update(cmd.Context(), args[0], content)
			if err != nil {
				return err
			}

			output.Println("Block updated successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "New content (Markdown)")
	cmd.MarkFlagRequired("content")
	return cmd
}

// NewAppendCmd creates the append command.
func NewAppendCmd() *cobra.Command {
	var content string
	cmd := &cobra.Command{
		Use:   "append <parent-id>",
		Short: "Append a block to a parent",
		Long:  "Append a new block as a child of the specified parent block.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			blockID, err := l.Append(cmd.Context(), args[0], content)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(map[string]string{"id": blockID})
			}

			output.Printf("Block appended successfully (ID: %s)\n", blockID)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "Block content (Markdown)")
	cmd.MarkFlagRequired("content")
	return cmd
}

// NewInsertAfterCmd creates the insert-after command.
func NewInsertAfterCmd() *cobra.Command {
	var content string
	cmd := &cobra.Command{
		Use:   "insert-after <previous-id>",
		Short: "Insert a block after a specific block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			blockID, err := l.InsertAfter(cmd.Context(), args[0], content)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(map[string]string{"id": blockID})
			}

			output.Printf("Block inserted successfully (ID: %s)\n", blockID)
			return nil
		},
	}
	cmd.Flags().StringVar(&content, "content", "", "Block content (Markdown)")
	cmd.MarkFlagRequired("content")
	return cmd
}

// NewDeleteCmd creates the delete command.
func NewDeleteCmd() *cobra.Command {
	var yesFlag bool
	cmd := &cobra.Command{
		Use:   "delete <block-id>",
		Short: "Delete a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yesFlag {
				return fmt.Errorf("use --yes to confirm deletion")
			}

			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			err = l.Delete(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			output.Println("Block deleted successfully")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Confirm deletion")
	return cmd
}

// NewMoveCmd creates the move command.
func NewMoveCmd() *cobra.Command {
	var previousID, parentID string
	cmd := &cobra.Command{
		Use:   "move <block-id>",
		Short: "Move a block to a new position",
		Long:  "Move a block to a new position. Specify either --previous-id or --parent-id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if previousID == "" && parentID == "" {
				return fmt.Errorf("either --previous-id or --parent-id must be specified")
			}

			l, err := logic.NewBlockLogic()
			if err != nil {
				return err
			}

			err = l.Move(cmd.Context(), args[0], previousID, parentID)
			if err != nil {
				return err
			}

			output.Println("Block moved successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&previousID, "previous-id", "", "ID of the block to insert after")
	cmd.Flags().StringVar(&parentID, "parent-id", "", "ID of the parent block to append to")
	return cmd
}

// NewBlockCmd creates the block command group.
func NewBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block operations",
		Long:  "Commands for managing SiYuan blocks (content blocks).",
	}

	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewChildrenCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewAppendCmd())
	cmd.AddCommand(NewInsertAfterCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewMoveCmd())

	return cmd
}
