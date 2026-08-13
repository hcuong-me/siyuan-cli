package tool

import "context"

func runStatus(ctx context.Context, request Request) Response {
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "status", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}
	switch request.Operation {
	case "version":
		value, err := client.GetVersion(ctx)
		if err != nil {
			return remoteResponse("status", err)
		}
		return success("status", "server version", map[string]string{"version": value})
	case "time":
		value, err := client.GetCurrentTime(ctx)
		if err != nil {
			return remoteResponse("status", err)
		}
		return success("status", "server time", map[string]int64{"time": value})
	case "boot_progress":
		value, err := client.GetBootProgress(ctx)
		if err != nil {
			return remoteResponse("status", err)
		}
		return success("status", "boot progress", value)
	default:
		return errorResponse("status", InvalidRequest, "unknown operation", "run tools to inspect status operations", false)
	}
}
