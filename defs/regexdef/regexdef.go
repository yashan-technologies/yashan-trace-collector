package regexdef

import "regexp"

const (
	// 24h
	range_format = `^[1-9][0-9]*[yMdhms]$`
	// yyyy-MM-dd-hh-mm
	time_format = `^[1-9]\d{3}-(0\d|1[0-2])-([012]\d|3[01])(-(0\d|1\d|2[0-3])(-([0-5]\d))?)?$`
	// 绝对路径 | 相对路径
	path_format = `^(\/?(\.{1,2}|[^\s]+)\/?([^\s]+\/)*[^\s]*)$`

	yasdb_process_format = `^[/\w._-]*yasdb[\0](?i)(open|nomount|mount)[\0]-D[\0][/\w._-]+[\0]`

	space_format     = `\s+`
	key_value_format = `^([^=]+)=(.*)$`
)

var (
	RangeRegex        = regexp.MustCompile(range_format)
	TimeRegex         = regexp.MustCompile(time_format)
	PathRegex         = regexp.MustCompile(path_format)
	SpaceRegex        = regexp.MustCompile(space_format)
	YasdbProcessRegex = regexp.MustCompile(yasdb_process_format)
	KeyValueRegex     = regexp.MustCompile(key_value_format)
)
