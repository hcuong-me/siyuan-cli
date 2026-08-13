// Package human adapts the agent tool protocol to a flag-based command line.
package human

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"siyuan/internal/tool"
)

// Execute runs the flag-based interface when the arguments match a cataloged
// operation, and otherwise delegates unchanged to the agent tool dispatcher.
func Execute(args []string, stdin io.Reader, stdout io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 1
	}
	toolName := args[0]
	if toolName == "-h" || toolName == "--help" {
		printUsage(stdout)
		return 0
	}
	if toolName == "tools" {
		return tool.Execute(args, stdin, stdout)
	}
	catalog := tool.Catalog()
	schema, ok := catalog[toolName]
	if !ok {
		return tool.Execute(args, stdin, stdout)
	}
	if len(args) >= 2 && (args[1] == "-h" || args[1] == "--help") {
		printToolUsage(stdout, toolName, schema)
		return 0
	}
	if len(args) >= 2 {
		if operationSchema, ok := schema.Operations[args[1]]; ok {
			return runFlagPath(toolName, args[1], operationSchema, args[2:], stdin, stdout)
		}
	}
	// A tool name alone is the agent path when JSON is piped on stdin. On an
	// interactive terminal there is no JSON stream to read: show the operation
	// list instead of blocking on input.
	if len(args) == 1 && interactiveStdin(stdin) {
		printToolUsage(stdout, toolName, schema)
		return 1
	}
	return tool.Execute(args, stdin, stdout)
}

func runFlagPath(toolName, operationName string, schema tool.OperationSchema, args []string, stdin io.Reader, stdout io.Writer) int {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printOperationUsage(stdout, toolName, operationName, schema)
			return 0
		}
	}
	parsed, err := buildRequest(toolName, operationName, schema, args)
	if err != nil {
		_, _ = fmt.Fprintf(stdout, "Error: %v\n", err)
		return 1
	}
	if parsed.json {
		if parsed.yes {
			return writeJSONFlagError(toolName, "--json cannot be combined with --yes; JSON side effects are preview-only", stdout)
		}
		if schema.SideEffect {
			if parsed.request.Mode == "apply" {
				return writeJSONFlagError(toolName, "--json cannot apply side effects; use text mode or an agent request with mode apply", stdout)
			}
			if parsed.request.Mode == "" {
				parsed.request.Mode = "preview"
			}
		}
		return executeEnvelope(toolName, parsed.request, stdout)
	}
	if !schema.SideEffect {
		response := execute(toolName, parsed.request)
		_, _ = fmt.Fprint(stdout, Render(response, operationName))
		return exitCode(response)
	}
	return runSideEffect(toolName, operationName, parsed, stdin, stdout)
}

func writeJSONFlagError(toolName, message string, stdout io.Writer) int {
	response := tool.Response{
		Version:     tool.ProtocolVersion,
		Tool:        toolName,
		OK:          false,
		Summary:     message,
		Warnings:    []tool.Warning{},
		NextActions: []tool.NextAction{{Tool: "tools", Reason: "use JSON only for read-only requests or preview side effects"}},
		Error: &tool.Error{
			Code:      tool.InvalidRequest,
			Message:   message,
			Fix:       "remove --yes and --mode apply, or use the agent JSON path for explicit apply",
			Retryable: false,
		},
	}
	if err := tool.WriteResponse(stdout, response); err != nil {
		return 1
	}
	return 1
}

func runSideEffect(toolName, operationName string, parsed parsedFlags, stdin io.Reader, stdout io.Writer) int {
	if parsed.request.Mode == "apply" {
		response := execute(toolName, parsed.request)
		_, _ = fmt.Fprint(stdout, Render(response, operationName))
		return exitCode(response)
	}
	previewRequest := parsed.request
	previewRequest.Mode = "preview"
	previewRequest.ConfirmationToken = ""
	preview := execute(toolName, previewRequest)
	if !preview.OK {
		_, _ = fmt.Fprint(stdout, Render(preview, operationName))
		return exitCode(preview)
	}
	_, _ = fmt.Fprint(stdout, Render(preview, operationName))
	token := confirmationToken(preview)
	if token == "" {
		_, _ = fmt.Fprintln(stdout, "Error: preview returned no confirmation token")
		return 1
	}
	if parsed.yes {
		return applyWithToken(toolName, operationName, parsed.request, token, stdout)
	}
	if !interactiveStdin(stdin) {
		_, _ = fmt.Fprintln(stdout, "Refusing to apply: stdin is not interactive. Re-run with --yes to confirm.")
		return 1
	}
	_, _ = fmt.Fprint(stdout, "Apply? [y/N] ")
	answer, _ := bufio.NewReader(stdin).ReadString('\n')
	if !isAffirmative(answer) {
		_, _ = fmt.Fprintln(stdout, "Cancelled.")
		return 0
	}
	return applyWithToken(toolName, operationName, parsed.request, token, stdout)
}

func applyWithToken(toolName, operationName string, request tool.Request, token string, stdout io.Writer) int {
	request.Mode = "apply"
	request.ConfirmationToken = token
	response := execute(toolName, request)
	_, _ = fmt.Fprint(stdout, Render(response, operationName))
	return exitCode(response)
}

// executeEnvelope writes the raw agent envelope for a --json request so the
// output is byte-identical to the agent path.
func executeEnvelope(toolName string, request tool.Request, stdout io.Writer) int {
	var stdin bytes.Buffer
	_ = json.NewEncoder(&stdin).Encode(request)
	return tool.Execute([]string{toolName}, &stdin, stdout)
}

// execute runs one request through the tool layer and decodes the response.
func execute(toolName string, request tool.Request) tool.Response {
	var stdin bytes.Buffer
	_ = json.NewEncoder(&stdin).Encode(request)
	var stdout bytes.Buffer
	_ = tool.Execute([]string{toolName}, &stdin, &stdout)
	var response tool.Response
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return tool.Response{
			Version: tool.ProtocolVersion,
			Tool:    toolName,
			OK:      false,
			Summary: "the tool returned no readable response",
			Error: &tool.Error{
				Code:      tool.RemoteError,
				Message:   "the tool returned no readable response",
				Fix:       "retry the request",
				Retryable: true,
			},
		}
	}
	return response
}

func exitCode(response tool.Response) int {
	if response.OK {
		return 0
	}
	return 1
}

func confirmationToken(response tool.Response) string {
	data, ok := response.Data.(map[string]any)
	if !ok {
		return ""
	}
	confirmation, ok := data["confirmation"].(map[string]any)
	if !ok {
		return ""
	}
	token, _ := confirmation["token"].(string)
	return token
}

func printUsage(stdout io.Writer) {
	_, _ = fmt.Fprint(stdout, `siyuan-cli — SiYuan Note command line

Usage:
  siyuan-cli tools                          show the agent tool catalog
  siyuan-cli <tool> <operation> [flags]     run an operation with flags

Tools:
  status     read server state
  context    find and read note data
  write      change document, block, and attribute content
  organize   change notebook, document, tag, and block structure
  export     preview or create document exports
  maintain   manage templates, assets, snapshots, and raw files

Run "siyuan-cli <tool> -h" for an operation list.
`)
}

func printToolUsage(stdout io.Writer, toolName string, schema tool.Schema) {
	_, _ = fmt.Fprintf(stdout, "siyuan-cli %s — %s\n\nOperations:\n", toolName, schema.Goal)
	for _, name := range sortedOperationNames(schema) {
		operation := schema.Operations[name]
		marker := "read"
		if operation.SideEffect {
			marker = "side effect"
		}
		_, _ = fmt.Fprintf(stdout, "  %-22s %s [%s]\n", name, operation.Description, marker)
	}
	_, _ = fmt.Fprintf(stdout, "\nRun \"siyuan-cli %s <operation> -h\" for flags.\n", toolName)
}

func printOperationUsage(stdout io.Writer, toolName, operationName string, schema tool.OperationSchema) {
	kind := "read-only"
	if schema.SideEffect {
		kind = "side effect (preview before apply)"
	}
	_, _ = fmt.Fprintf(stdout, "Usage: siyuan-cli %s %s [flags]\n\n", toolName, operationName)
	_, _ = fmt.Fprintf(stdout, "%s\n\n", schema.Description)
	_, _ = fmt.Fprintf(stdout, "Type: %s\n", kind)
	if schema.Risk != "" {
		_, _ = fmt.Fprintf(stdout, "Risk boundary: %s\n", schema.Risk)
	}
	_, _ = fmt.Fprintf(stdout, "\nFlags:\n")
	for _, name := range sortedFieldNames(schema.Input) {
		field := schema.Input[name]
		required := ""
		if field.Required {
			required = " (required)"
		}
		_, _ = fmt.Fprintf(stdout, "  --%-20s %s%s\n", name, field.Type, required)
	}
	if schema.SideEffect {
		_, _ = fmt.Fprintf(stdout, "  --%-20s preview or apply (default: preview)\n", "mode")
		_, _ = fmt.Fprintf(stdout, "  --%-20s token returned by preview (for --mode apply)\n", "confirmation_token")
		_, _ = fmt.Fprintf(stdout, "  --%-20s confirm the preview and apply without prompting\n", "yes")
	}
	_, _ = fmt.Fprintf(stdout, "  --%-20s print the raw agent JSON envelope\n", "json")
}

func sortedOperationNames(schema tool.Schema) []string {
	names := make([]string, 0, len(schema.Operations))
	for name := range schema.Operations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedFieldNames(input map[string]tool.FieldSchema) []string {
	names := make([]string, 0, len(input))
	for name := range input {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
