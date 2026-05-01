// Package output provides utilities for formatting CLI output.
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

// AsJSON outputs data as formatted JSON.
func AsJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// AsJSONRaw outputs raw JSON bytes.
func AsJSONRaw(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return AsJSON(v)
}

// AsTable outputs data as a formatted table.
func AsTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}

// Error prints an error message to stderr.
func Error(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// Println prints a line to stdout.
func Println(a ...interface{}) {
	fmt.Println(a...)
}

// Printf prints formatted output to stdout.
func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
