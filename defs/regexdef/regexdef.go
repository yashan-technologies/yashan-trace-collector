package regexdef

import "regexp"

const (
	// 24h
	range_format = `^[1-9][0-9]*[yMdhms]$`
	// yyyy-MM-dd-hh-mm
	time_format = `^[1-9]\d{3}-(0\d|1[0-2])-([012]\d|3[01])(-(0\d|1\d|2[0-3])(-([0-5]\d))?)?$`
	// 绝对路径 | 相对路径
	path_format = `^((/[a-zA-Z0-9_-]+)+|([a-zA-Z0-9_-]|./[a-zA-Z0-9_-])*(/[a-zA-Z0-9_-]+)*/?)$`

	yasdb_process_format = `^[/\w._-]*yasdb[\0](?i)(open|nomount|mount)[\0]-D[\0][/\w._-]+[\0]`

	space_format = `\s+`
)

var (
	RANGE_REGEX         = regexp.MustCompile(range_format)
	TIME_REGEX          = regexp.MustCompile(time_format)
	PATH_REGEX          = regexp.MustCompile(path_format)
	SPACE_REGEX         = regexp.MustCompile(space_format)
	YASDB_PROCESS_REGEX = regexp.MustCompile(yasdb_process_format)
)
