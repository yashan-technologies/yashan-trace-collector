package commons

import "git.yasdb.com/go/yaserr"

// ParseString converts the data to a string value, and returns an error with name and context if failed.
func ParseString(name string, data interface{}, context string) (value string, err error) {
	value, ok := data.(string)
	if !ok {
		err = &ErrInterfaceTypeNotMatch{
			Key:     name,
			Targets: []interface{}{""},
			Current: data,
		}
		err = yaserr.Wrapf(err, context)
		return
	}
	return
}

// ParseBool converts the data to a bool value, and returns an error with name and context if failed.
func ParseBool(name string, data interface{}, context string) (value bool, err error) {
	value, ok := data.(bool)
	if !ok {
		err = &ErrInterfaceTypeNotMatch{
			Key:     name,
			Targets: []interface{}{false},
			Current: data,
		}
		err = yaserr.Wrapf(err, context)
		return
	}
	return
}
