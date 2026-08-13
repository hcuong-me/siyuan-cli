package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

var jsonUnmarshal = json.Unmarshal

// decodeInput decodes an already schema-validated input object while keeping
// JSON numbers as json.Number. This prevents handlers from observing a
// precision-losing float64 value after the protocol boundary.
func decodeInput(request Request) (map[string]any, error) {
	var shape map[string]json.RawMessage
	if err := jsonUnmarshal(request.Input, &shape); err != nil || shape == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}

	decoder := json.NewDecoder(bytes.NewReader(request.Input))
	decoder.UseNumber()

	var input map[string]any
	if err := decoder.Decode(&input); err != nil {
		return nil, fmt.Errorf("decode input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("input must contain one JSON object")
		}
		return nil, fmt.Errorf("read input: %w", err)
	}
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return input, nil
}

// stringInputStrict returns a present string value without coercion. A
// missing field is represented by an empty string and nil error so optional
// fields retain their v1 behavior; a present value of another JSON type is an
// INVALID_REQUEST error.
func stringInputStrict(input map[string]any, name string) (string, error) {
	value, ok := input[name]
	if !ok {
		return "", nil
	}
	decoded, ok := value.(string)
	if !ok {
		return "", invalidInputTypeError(name, "string")
	}
	return decoded, nil
}

// integerInputStrict returns a present integer without coercion. It accepts
// json.Number values produced by decodeInput and native integer values used by
// focused unit tests. Floating-point values are accepted only when finite,
// integral, and within the platform int range.
func integerInputStrict(input map[string]any, name string) (int, error) {
	value, ok := input[name]
	if !ok {
		return 0, nil
	}

	switch decoded := value.(type) {
	case json.Number:
		parsed, err := parseInteger([]byte(decoded.String()))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case int:
		return decoded, nil
	case int8:
		return int(decoded), nil
	case int16:
		return int(decoded), nil
	case int32:
		return int(decoded), nil
	case int64:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%d", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case uint:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%d", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case uint8:
		return int(decoded), nil
	case uint16:
		return int(decoded), nil
	case uint32:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%d", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case uint64:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%d", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case float32:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%g", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	case float64:
		parsed, err := parseInteger([]byte(fmt.Sprintf("%g", decoded)))
		if err != nil {
			return 0, invalidInputTypeError(name, "integer")
		}
		return parsed, nil
	default:
		return 0, invalidInputTypeError(name, "integer")
	}
}

// objectInputStrict returns a JSON object without coercing arrays or scalar
// values. decodeInput represents nested JSON objects as map[string]any.
func objectInputStrict(input map[string]any, name string) (map[string]any, error) {
	value, ok := input[name]
	if !ok {
		return nil, nil
	}
	decoded, ok := value.(map[string]any)
	if !ok {
		return nil, invalidInputTypeError(name, "object")
	}
	return decoded, nil
}

func stringInput(input map[string]any, name string) string {
	value, _ := stringInputStrict(input, name)
	return value
}

func integerInput(input map[string]any, name string, fallback int) int {
	value, err := integerInputStrict(input, name)
	if err != nil || value == 0 {
		return fallback
	}
	return value
}
