// Package file provides commands for file operations.
package file

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewTreeCmd creates the tree command.
func NewTreeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tree <path>",
		Short: "List directory contents",
		Long: `List files and directories at the given path.

Path can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file tree /assets
  siyuan file tree /data/assets
  siyuan file tree assets`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			files, err := l.Tree(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(files)
			}

			if len(files) == 0 {
				output.Println("Directory is empty.")
				return nil
			}

			output.Printf("Found %d items\n\n", len(files))

			rows := make([][]string, len(files))
			for i, f := range files {
				typ := "file"
				if f.IsDir {
					typ = "dir"
				}
				rows[i] = []string{typ, f.Name}
			}
			output.AsTable([]string{"TYPE", "NAME"}, rows)
			return nil
		},
	}
}

// NewReadCmd creates the read command.
func NewReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <path>",
		Short: "Read file content",
		Long: `Read the content of a file.

Path can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file read /assets/example.md
  siyuan file read /data/storage/notes.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			content, err := l.Read(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			// Output raw content (for piping)
			output.Println(content)
			return nil
		},
	}
}

// NewWriteCmd creates the write command.
func NewWriteCmd() *cobra.Command {
	var content string
	cmd := &cobra.Command{
		Use:   "write <path>",
		Short: "Write content to file",
		Long: `Write content to a file. Creates the file if it doesn't exist.

Path can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file write /data/storage/notes.txt --content "Hello world"
  echo "Hello" | siyuan file write /data/storage/greeting.txt`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			if content == "" {
				return fmt.Errorf("--content is required")
			}

			if err := l.Write(cmd.Context(), args[0], content, false); err != nil {
				return err
			}

			output.Println("File written successfully.")
			return nil
		},
	}
	cmd.Flags().StringVarP(&content, "content", "c", "", "Content to write")
	return cmd
}

// NewMkdirCmd creates the mkdir command.
func NewMkdirCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mkdir <path>",
		Short: "Create a directory",
		Long: `Create a new directory.

Path can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file mkdir /data/storage/new-folder`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			if err := l.Write(cmd.Context(), args[0], "", true); err != nil {
				return err
			}

			output.Println("Directory created successfully.")
			return nil
		},
	}
}

// NewRemoveCmd creates the remove command.
func NewRemoveCmd() *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a file or directory",
		Long: `Remove a file or directory.

Path can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file remove /data/storage/old.txt --yes
  siyuan file remove /data/temp --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("use --yes to confirm removal")
			}

			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			if err := l.Remove(cmd.Context(), args[0]); err != nil {
				return err
			}

			output.Println("Removed successfully.")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm removal")
	return cmd
}

// NewRenameCmd creates the rename command.
func NewRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-path> <new-path>",
		Short: "Rename or move a file",
		Long: `Rename or move a file or directory.

Both paths can be relative to /data/ or absolute starting with /data/.

Examples:
  siyuan file rename old.txt new.txt
  siyuan file rename /data/temp/file.txt /data/storage/file.txt`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewFileLogic()
			if err != nil {
				return err
			}

			if err := l.Rename(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			output.Printf("Renamed %s -> %s\n", filepath.Base(args[0]), filepath.Base(args[1]))
			return nil
		},
	}
}

// NewFileCmd creates the file command group.
func NewFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "File operations",
		Long:  "Commands for managing files in SiYuan data directory (raw filesystem access).",
	}

	cmd.AddCommand(NewTreeCmd())
	cmd.AddCommand(NewReadCmd())
	cmd.AddCommand(NewWriteCmd())
	cmd.AddCommand(NewMkdirCmd())
	cmd.AddCommand(NewRemoveCmd())
	cmd.AddCommand(NewRenameCmd())

	return cmd
}
