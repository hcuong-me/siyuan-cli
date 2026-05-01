// Package attr provides commands for attribute operations.
package attr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewGetCmd creates the get command.
func NewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <block-id>",
		Short: "Get attributes of a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewAttrLogic()
			if err != nil {
				return err
			}

			attrs, err := l.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(attrs)
			}

			rows := make([][]string, 0, len(attrs))
			for key, value := range attrs {
				rows = append(rows, []string{key, value})
			}
			output.AsTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
	}
}

// NewSetCmd creates the set command.
func NewSetCmd() *cobra.Command {
	var key, value string
	cmd := &cobra.Command{
		Use:   "set <block-id>",
		Short: "Set an attribute on a block",
		Long:  "Set a custom attribute on a block. Custom attributes must be prefixed with 'custom-'.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				return fmt.Errorf("--key is required")
			}

			l, err := logic.NewAttrLogic()
			if err != nil {
				return err
			}

			err = l.SetSingle(cmd.Context(), args[0], key, value)
			if err != nil {
				return err
			}

			output.Printf("Attribute '%s' set on block %s\n", key, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Attribute key (required)")
	cmd.Flags().StringVar(&value, "value", "", "Attribute value")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// NewSetMultipleCmd creates the set-multiple command.
func NewSetMultipleCmd() *cobra.Command {
	var attrs []string
	cmd := &cobra.Command{
		Use:   "set-multiple <block-id>",
		Short: "Set multiple attributes on a block",
		Long:  "Set multiple attributes at once using key=value pairs. Example: --attr key1=value1 --attr key2=value2",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(attrs) == 0 {
				return fmt.Errorf("at least one --attr is required")
			}

			attrMap := make(map[string]string)
			for _, attr := range attrs {
				parts := strings.SplitN(attr, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid attribute format: %s (expected key=value)", attr)
				}
				attrMap[parts[0]] = parts[1]
			}

			l, err := logic.NewAttrLogic()
			if err != nil {
				return err
			}

			err = l.Set(cmd.Context(), args[0], attrMap)
			if err != nil {
				return err
			}

			output.Printf("Attributes set on block %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&attrs, "attr", []string{}, "Attribute key=value pair (can be specified multiple times)")
	return cmd
}

// NewResetCmd creates the reset command.
func NewResetCmd() *cobra.Command {
	var key string
	cmd := &cobra.Command{
		Use:   "reset <block-id>",
		Short: "Remove an attribute from a block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				return fmt.Errorf("--key is required")
			}

			l, err := logic.NewAttrLogic()
			if err != nil {
				return err
			}

			err = l.Reset(cmd.Context(), args[0], key)
			if err != nil {
				return err
			}

			output.Printf("Attribute '%s' removed from block %s\n", key, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Attribute key to remove (required)")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

// NewBookmarksCmd creates the bookmarks command.
func NewBookmarksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bookmarks",
		Short: "List all bookmark labels",
		RunE: func(cmd *cobra.Command, _ []string) error {
			l, err := logic.NewAttrLogic()
			if err != nil {
				return err
			}

			labels, err := l.GetBookmarkLabels(cmd.Context())
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(labels)
			}

			rows := make([][]string, len(labels))
			for i, label := range labels {
				rows[i] = []string{label.Label, fmt.Sprintf("%d", label.Count)}
			}
			output.AsTable([]string{"LABEL", "COUNT"}, rows)
			return nil
		},
	}
}

// NewAttrCmd creates the attr command group.
func NewAttrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attr",
		Aliases: []string{"attribute"},
		Short:   "Attribute operations",
		Long:    "Commands for managing block attributes in SiYuan.",
	}

	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewSetCmd())
	cmd.AddCommand(NewSetMultipleCmd())
	cmd.AddCommand(NewResetCmd())
	cmd.AddCommand(NewBookmarksCmd())

	return cmd
}
