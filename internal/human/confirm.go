package human

import (
	"io"
	"os"
	"strings"
)

// interactiveStdin reports whether the given stdin is an interactive terminal.
// Tests override it to exercise the interactive and non-interactive paths.
var interactiveStdin = func(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}
	stat, err := file.Stat()
	return err == nil && stat.Mode()&os.ModeCharDevice != 0
}

// isAffirmative reports whether the user answered yes to the apply prompt.
func isAffirmative(answer string) bool {
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
