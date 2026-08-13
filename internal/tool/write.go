package tool

import (
	"context"

	"siyuan/internal/siyuan"
)

func runWrite(ctx context.Context, request Request) Response {
	input, err := decodeInput(request)
	if err != nil {
		return errorResponse("write", InvalidRequest, "input is not valid JSON", "send an input object", false)
	}
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "write", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}
	return sideEffect("write", request, input, func() ([]Target, error) { return stateTargets(ctx, client, request, input) }, func() (any, error) {
		switch request.Operation {
		case "create_document":
			notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
			if err != nil {
				return nil, err
			}
			created, err := client.CreateDocumentWithMarkdownResult(ctx, notebook.ID, stringInput(input, "path"), stringInput(input, "content"))
			result := map[string]string{"id": created.ID, "notebook": notebook.ID}
			if created.Box != "" {
				result["box"] = created.Box
			}
			if created.Path != "" {
				result["path"] = created.Path
			}
			if created.HPath != "" {
				result["hPath"] = created.HPath
			}
			return result, err
		case "update_document":
			resolution, err := resolveDocumentForRequest(ctx, client, request, stringInput(input, "notebook"), stringInput(input, "path"))
			if err != nil {
				return nil, err
			}
			err = client.UpdateBlock(ctx, resolution.ID, "markdown", stringInput(input, "content"))
			return map[string]string{"id": resolution.ID, "notebook": resolution.Notebook, "path": resolution.Path}, err
		case "update_block":
			err := client.UpdateBlock(ctx, stringInput(input, "block_id"), "markdown", stringInput(input, "content"))
			return map[string]string{"id": stringInput(input, "block_id")}, err
		case "append_block":
			id, err := client.InsertBlock(ctx, "markdown", stringInput(input, "content"), "", "", stringInput(input, "parent_id"))
			return map[string]string{"id": id}, err
		case "insert_after_block":
			id, err := client.InsertBlock(ctx, "markdown", stringInput(input, "content"), "", stringInput(input, "previous_id"), "")
			return map[string]string{"id": id}, err
		case "set_attribute":
			err := client.SetBlockAttrs(ctx, stringInput(input, "block_id"), map[string]string{stringInput(input, "key"): stringInput(input, "value")})
			return map[string]string{"id": stringInput(input, "block_id")}, err
		case "set_attributes":
			attrs := map[string]string{}
			object, err := objectInputStrict(input, "attributes")
			if err != nil {
				return nil, err
			}
			for key, value := range object {
				text, ok := value.(string)
				if !ok {
					return nil, invalidInputTypeError("attributes."+key, "string")
				}
				attrs[key] = text
			}
			err = client.SetBlockAttrs(ctx, stringInput(input, "block_id"), attrs)
			return map[string]string{"id": stringInput(input, "block_id")}, err
		case "reset_attribute":
			err := client.ResetBlockAttr(ctx, stringInput(input, "block_id"), stringInput(input, "key"))
			return map[string]string{"id": stringInput(input, "block_id")}, err
		default:
			return nil, fmtNotFound("operation")
		}
	})
}

func fmtNotFound(item string) error { return &notFoundError{item: item} }

type notFoundError struct{ item string }

// Error returns the missing item message.
func (e *notFoundError) Error() string { return e.item + " not found" }

var _ = siyuan.BlockAttrs{}
