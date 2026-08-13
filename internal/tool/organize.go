package tool

import "context"

func runOrganize(ctx context.Context, request Request) Response {
	input, err := decodeInput(request)
	if err != nil {
		return errorResponse("organize", InvalidRequest, "input is not valid JSON", "send an input object", false)
	}
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "organize", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}
	return sideEffect("organize", request, input, func() ([]Target, error) { return stateTargets(ctx, client, request, input) }, func() (any, error) {
		switch request.Operation {
		case "create_notebook":
			value, err := client.CreateNotebook(ctx, stringInput(input, "name"))
			return value, err
		case "rename_notebook":
			notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
			if err != nil {
				return nil, err
			}
			err = client.RenameNotebook(ctx, notebook.ID, stringInput(input, "name"))
			return map[string]string{"id": notebook.ID, "name": stringInput(input, "name")}, err
		case "open_notebook":
			notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
			if err != nil {
				return nil, err
			}
			err = client.OpenNotebook(ctx, notebook.ID)
			return map[string]string{"id": notebook.ID, "name": notebook.Name}, err
		case "close_notebook":
			notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
			if err != nil {
				return nil, err
			}
			err = client.CloseNotebook(ctx, notebook.ID)
			return map[string]string{"id": notebook.ID, "name": notebook.Name}, err
		case "remove_notebook":
			notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
			if err != nil {
				return nil, err
			}
			err = client.RemoveNotebook(ctx, notebook.ID)
			return map[string]string{"id": notebook.ID, "name": notebook.Name}, err
		case "move_block":
			err := client.MoveBlock(ctx, stringInput(input, "block_id"), stringInput(input, "previous_id"), stringInput(input, "parent_id"))
			result := map[string]string{"id": stringInput(input, "block_id"), "parent_id": stringInput(input, "parent_id")}
			if previous := stringInput(input, "previous_id"); previous != "" {
				result["previous_id"] = previous
			}
			return result, err
		case "delete_block":
			err := client.DeleteBlock(ctx, stringInput(input, "block_id"))
			return map[string]string{"id": stringInput(input, "block_id")}, err
		case "remove_document":
			resolution, err := resolveDocumentForRequest(ctx, client, request, stringInput(input, "notebook"), stringInput(input, "path"))
			if err != nil {
				return nil, err
			}
			err = client.RemoveDocumentByID(ctx, resolution.ID)
			return map[string]string{"id": resolution.ID, "notebook": resolution.Notebook, "path": resolution.Path}, err
		case "rename_tag":
			tag, err := resolveTagForRequest(ctx, client, request, stringInput(input, "tag"))
			if err != nil {
				return nil, err
			}
			err = client.RenameTag(ctx, tag.Label, stringInput(input, "name"))
			return map[string]any{"tag": tag.Label, "count": tag.Count}, err
		case "remove_tag":
			tag, err := resolveTagForRequest(ctx, client, request, stringInput(input, "tag"))
			if err != nil {
				return nil, err
			}
			err = client.RemoveTag(ctx, tag.Label)
			return map[string]any{"tag": tag.Label, "count": tag.Count}, err
		default:
			return nil, fmtNotFound("operation")
		}
	})
}
