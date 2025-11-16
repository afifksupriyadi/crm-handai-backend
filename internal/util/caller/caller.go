package caller

import (
	"fmt"
	"runtime"
)

// GetWithDepth returns the file path and line number of the caller at the specified depth.
// Depth 0 is the caller of GetWithDepth, depth 1 is the caller's caller, and so on.
// Returns an empty string if the caller information cannot be retrieved.
func GetWithDepth(depth int) string {
	if depth < 0 {
		return ""
	}
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}
