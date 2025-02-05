package pwdutil

import (
	"regexp"
)

// PrintLogDesensitize hides the password.
func PrintLogDesensitize(args string) string {
	type container struct {
		expr string
		repl string
	}
	rmap := make(map[string]container)
	// yasql command
	rmap["yasql command"] = container{
		expr: `(yasql\s+\S+/)"(.+?)"(\s+)`,
		repl: "${1}******${3}",
	}
	// create user command
	rmap["create user command"] = container{
		expr: `(?i)(user)(.+)(identified by )(.+)`,
		repl: "${1}${2}${3}******;",
	}

	// yasboot command
	rmap["yasboot command"] = container{
		expr: `(yasboot sql)(.+)(--password)(.+)(--sql)`,
		repl: "${1}${2}${3} ****** ${5}",
	}

	// replace all
	for _, r := range rmap {
		args = regexpReplace(r.expr, args, r.repl)
	}
	return args
}

func regexpReplace(expr, input, repl string) string {
	r, err := regexp.Compile(expr)
	if err != nil {
		return input
	}
	return r.ReplaceAllString(input, repl)
}
