package errdef

import (
	"fmt"
	"strings"

	"ytc/utils/stringutil"
)

type ErrYtcFlag struct {
	Flag     string
	Value    string
	Examples []string
	Help     string
}

func NewErrYtcFlag(flag, value string, examples []string, help string) *ErrYtcFlag {
	return &ErrYtcFlag{
		Flag:     flag,
		Value:    value,
		Examples: examples,
		Help:     help,
	}
}

func (e ErrYtcFlag) Error() string {
	var wrapExamples []string
	for _, e := range e.Examples {
		wrapExamples = append(wrapExamples, fmt.Sprintf("'%s'", e))
	}
	message := fmt.Sprintf("The value of %s: %s is invalid, the available input formats are as follows: [%s]", e.Flag, e.Value, strings.Join(wrapExamples, stringutil.STR_COMMA))
	if len(e.Help) != 0 {
		message += stringutil.STR_COMMA + e.Help
	}
	return message
}
