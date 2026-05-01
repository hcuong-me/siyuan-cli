// Package sqlcmd provides commands for SQL query operations.
package sqlcmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"siyuan/internal/logic"
	"siyuan/internal/utils/output"
)

// NewQueryCmd creates the query command.
func NewQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query <sql-statement>",
		Short: "Execute a SQL query on the SiYuan database",
		Long: `Execute a SQL query on the SiYuan database.

SECURITY WARNING:
  Only SELECT queries are allowed. The following operations are blocked:
  - DELETE, DROP, TRUNCATE
  - UPDATE, INSERT, REPLACE
  - CREATE, ALTER
  - GRANT, REVOKE
  - And other data modification commands

Examples:
  siyuan sql query "SELECT * FROM blocks LIMIT 10"
  siyuan sql query "SELECT id, content, type FROM blocks WHERE type='h'"
  siyuan sql query "SELECT COUNT(*) as total FROM blocks" --json

Available tables:
  - blocks: Main content blocks
  - notebooks: Notebook information
  - refs: Block references
  - attributes: Block attributes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			l, err := logic.NewSQLLogic()
			if err != nil {
				return err
			}

			query := strings.TrimSpace(args[0])

			result, err := l.Query(cmd.Context(), query)
			if err != nil {
				return err
			}

			if cmd.Flag("json").Changed {
				return output.AsJSON(result)
			}

			if len(result) == 0 {
				output.Println("No results found.")
				return nil
			}

			// Print row count
			output.Printf("Found %d rows\n\n", len(result))

			// Get headers from first row
			firstRow := result[0]
			headers := make([]string, 0, len(firstRow))
			for key := range firstRow {
				headers = append(headers, key)
			}

			// Build rows
			rows := make([][]string, len(result))
			for i, row := range result {
				rowData := make([]string, len(headers))
				for j, header := range headers {
					if val, ok := row[header]; ok {
						rowData[j] = fmt.Sprintf("%v", val)
					} else {
						rowData[j] = ""
					}
				}
				rows[i] = rowData
			}

			output.AsTable(headers, rows)
			return nil
		},
	}
}

// NewSQLCmd creates the sql command group.
func NewSQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql",
		Short: "SQL operations",
		Long:  "Commands for executing SQL queries on the SiYuan database (read-only).",
	}

	cmd.AddCommand(NewQueryCmd())

	return cmd
}
