package commons

import (
	"fmt"
	"strings"
)

type ErrInterfaceTypeNotMatch struct {
	Key     string
	Targets []interface{}
	Current interface{}
}

func (e *ErrInterfaceTypeNotMatch) Error() string {
	prefix := fmt.Sprintf("data type of [%s] should be: ", e.Key)
	for _, t := range e.Targets {
		prefix += fmt.Sprintf("%T or ", t)
	}
	prefix = strings.TrimSuffix(prefix, " or ")
	return fmt.Sprintf("%s, not: %T", prefix, e.Current)
}
