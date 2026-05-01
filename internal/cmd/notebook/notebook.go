// Package notebook provides commands for notebook operations.
package notebook

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
		Short: "List all notebooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			notebooks, err := l.List(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(notebooks)
			}

			rows := make([][]string, len(notebooks))
			for i, nb := range notebooks {
				status := "open"
				if nb.Closed {
					status = "closed"
				}
				rows[i] = []string{nb.ID, nb.Name, status}
			}
			output.AsTable([]string{"ID", "NAME", "STATUS"}, rows)
			return nil
		},
	}
}

// NewCreateCmd creates the create command.
func NewCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			notebook, err := l.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(notebook)
			}

			output.Printf("Created notebook: %s (ID: %s)\n", notebook.Name, notebook.ID)
			return nil
		},
	}
}

// NewRenameCmd creates the rename command.
func NewRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <id> <name>",
		Short: "Rename a notebook",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			err = l.Rename(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}

			output.Println("Notebook renamed successfully")
			return nil
		},
	}
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	var yesFlag bool
	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove a notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yesFlag {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			err = l.Remove(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			output.Println("Notebook removed successfully")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yesFlag, "yes", false, "Confirm removal")
	return cmd
}

// NewOpenCmd creates the open command.
func NewOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <id>",
		Short: "Open a closed notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			err = l.Open(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			output.Println("Notebook opened successfully")
			return nil
		},
	}
}

// NewCloseCmd creates the close command.
func NewCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close <id>",
		Short: "Close an open notebook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewNotebookLogic()
			if err != nil {
				return err
			}

			err = l.Close(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			output.Println("Notebook closed successfully")
			return nil
		},
	}
}

// NewNotebookCmd creates the notebook command group.
func NewNotebookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notebook",
		Short: "Notebook operations",
		Long:  "Commands for managing SiYuan notebooks.",
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewRenameCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewOpenCmd())
	cmd.AddCommand(NewCloseCmd())

	return cmd
}
