package human

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"siyuan/internal/tool"
)

// parsedFlags is the result of converting command-line flags into a request.
type parsedFlags struct {
	request tool.Request
	json    bool
	yes     bool
}

// buildRequest converts --flag value arguments into one tool request. The
// reserved flags --json, --yes, --mode, and --confirmation_token map onto the
// request envelope; every other flag must be a field of the operation schema.
func buildRequest(toolName, operationName string, schema tool.OperationSchema, args []string) (parsedFlags, error) {
	input := map[string]any{}
	result := parsedFlags{request: tool.Request{Version: tool.ProtocolVersion, Operation: operationName}}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return result, fmt.Errorf("unexpected argument %q: use --flag value", arg)
		}
		name := strings.TrimPrefix(arg, "--")
		switch name {
		case "json":
			result.json = true
			continue
		case "yes":
			result.yes = true
			continue
		case "mode", "confirmation_token":
			value, next, err := nextFlagValue(args, i)
			if err != nil {
				return result, err
			}
			i = next
			if name == "mode" {
				result.request.Mode = value
			} else {
				result.request.ConfirmationToken = value
			}
			continue
		}
		field, ok := schema.Input[name]
		if !ok {
			return result, fmt.Errorf("unknown flag --%s for %s %s", name, toolName, operationName)
		}
		value, next, err := nextFlagValue(args, i)
		if err != nil {
			return result, err
		}
		i = next
		converted, err := convertFlagValue(name, field, value)
		if err != nil {
			return result, err
		}
		input[name] = converted
	}

	for name, field := range schema.Input {
		if field.Required {
			if _, ok := input[name]; !ok {
				return result, fmt.Errorf("missing required flag --%s", name)
			}
		}
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return result, fmt.Errorf("cannot build request input: %w", err)
	}
	result.request.Input = encoded
	return result, nil
}

func nextFlagValue(args []string, index int) (string, int, error) {
	flag := args[index]
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("flag %s requires a value", flag)
	}
	return args[index+1], index + 1, nil
}

func convertFlagValue(name string, field tool.FieldSchema, value string) (any, error) {
	switch field.Type {
	case "integer":
		number, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("flag --%s expects an integer, got %q", name, value)
		}
		return number, nil
	case "object":
		var object map[string]any
		if err := json.Unmarshal([]byte(value), &object); err != nil || object == nil {
			return nil, fmt.Errorf("flag --%s expects a JSON object, got %q", name, value)
		}
		return object, nil
	default:
		return value, nil
	}
}
