// The stringutil package encapsulates functions related to strings.
package stringutil

const (
	STR_EMPTY       = ""
	STR_BLANK_SPACE = " "
	STR_NEWLINE     = "\n"
	STR_COMMA       = ","
	STR_DOT         = "."
)

// IsEmpty checks whether a string is empty.
func IsEmpty(str string) bool {
	return len(str) == 0
}
