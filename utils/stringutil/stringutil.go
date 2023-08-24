// The stringutil package encapsulates functions related to strings.
package stringutil

import "regexp"

const (
	STR_EMPTY         = ""
	STR_BLANK_SPACE   = " "
	STR_NEWLINE       = "\n"
	STR_COMMA         = ","
	STR_DOT           = "."
	STR_HYPHEN        = "-"
	STR_BAR           = "|"
	STR_FORWARD_SLASH = "/"
	STR_UNDER_SCORE   = "_"
	STR_HASH          = "#"
	STR_HTML_BR       = "<br>"
	STR_QUESTION_MARK = "?"
)

// IsEmpty checks whether a string is empty.
func IsEmpty(str string) bool {
	return len(str) == 0
}

func RemoveExtraSpaces(str string) string {
	regex := regexp.MustCompile(`\s+`)
	return regex.ReplaceAllString(str, STR_BLANK_SPACE)
}
