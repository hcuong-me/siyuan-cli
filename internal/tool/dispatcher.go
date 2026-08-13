package tool

import (
	"context"
	"io"
	"strings"

	"siyuan/internal/siyuan"
)

var newClient = siyuan.New

// Execute dispatches one top-level tool and writes exactly one response.
func Execute(args []string, stdin io.Reader, stdout io.Writer) (exitCode int) {
	toolName := ""
	if len(args) > 0 {
		toolName = args[0]
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = WriteResponse(stdout, RecoverResponse(toolName, recovered))
			exitCode = 1
		}
	}()

	if len(args) != 1 {
		_ = WriteResponse(stdout, errorResponse(toolName, InvalidRequest, "provide exactly one top-level tool", "use tools, status, context, write, organize, export, or maintain", false))
		return 1
	}
	if toolName == "tools" || toolName == "-h" || toolName == "--help" {
		_ = WriteResponse(stdout, ToolsResponse())
		return 0
	}
	if _, ok := Catalog()[toolName]; !ok {
		_ = WriteResponse(stdout, errorResponse(toolName, InvalidRequest, "unknown top-level tool", "run tools to inspect the available tools", false))
		return 1
	}

	request, err := DecodeRequest(stdin)
	if err != nil {
		_ = WriteResponse(stdout, errorResponse(toolName, InvalidRequest, "invalid JSON request", "send one JSON object that matches the tool schema", false))
		return 1
	}
	if err := ValidateRequest(toolName, request); err != nil {
		_ = WriteResponse(stdout, errorResponse(toolName, InvalidRequest, err.Error(), "run tools and correct the request", false))
		return 1
	}

	response := dispatch(context.Background(), toolName, request)
	response = enrichResponse(response, toolName, request.Operation)
	if err := WriteResponse(stdout, response); err != nil {
		return 1
	}
	if !response.OK {
		return 1
	}
	return 0
}

// enrichResponse adds metadata that is safe to derive from the protocol
// boundary. Handler-specific selectors may add candidates themselves; this
// helper only supplies deterministic documentation and preview/apply actions
// when the operation is known.
func enrichResponse(response Response, toolName, operation string) Response {
	if response.Warnings == nil {
		response.Warnings = []Warning{}
	}
	if response.NextActions == nil {
		response.NextActions = []NextAction{}
	}
	if response.Error == nil {
		if response.OK && operation != "" {
			if schema, ok := Catalog()[toolName].Operations[operation]; ok && schema.SideEffect && response.Data != nil {
				data, isMap := response.Data.(map[string]any)
				_, hasPreview := data["preview"]
				if isMap && hasPreview {
					response.NextActions = []NextAction{{Tool: toolName, Operation: operation, Reason: "review irreversible_effects, then apply with the returned confirmation token"}}
				}
			}
			if operation == "search_documents" || operation == "search_blocks" {
				response.NextActions = []NextAction{{Tool: "context", Operation: "read_document", Reason: "use a returned stable document ID and canonical path for follow-up context"}}
			}
		}
		return response
	}
	response.Error = normalizeError(response.Error)
	if len(response.NextActions) == 0 {
		response.NextActions = defaultNextActions(toolName, operation, response.Error.Code)
	}
	return response
}

func dispatch(ctx context.Context, toolName string, request Request) Response {
	switch toolName {
	case "status":
		return runStatus(ctx, request)
	case "context":
		return runContext(ctx, request)
	case "write":
		return runWrite(ctx, request)
	case "organize":
		return runOrganize(ctx, request)
	case "export":
		return runExport(ctx, request)
	case "maintain":
		return runMaintain(ctx, request)
	default:
		return errorResponse(toolName, InvalidRequest, "unknown top-level tool", "run tools to inspect the available tools", false)
	}
}

func clientOrError() (*siyuan.Client, *Error) {
	client, err := newClient()
	if err != nil {
		return nil, normalizeError(&Error{Code: MissingConfig, Message: "SiYuan configuration is missing", Fix: "set SIYUAN_TOKEN and, if needed, SIYUAN_BASE_URL", Retryable: false})
	}
	return client, nil
}

func remoteResponse(toolName string, err error) Response {
	message := err.Error()
	code := RemoteError
	fix := "check that SiYuan is running and retry the request"
	lowerMessage := strings.ToLower(message)
	if strings.Contains(lowerMessage, "not found") || strings.Contains(lowerMessage, "does not exist") || strings.Contains(lowerMessage, "no such") {
		code = NotFound
		fix = "check the selector or ID and retry"
	}
	return errorResponse(toolName, code, message, fix, true)
}
