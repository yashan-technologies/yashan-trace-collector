package errdef

import "fmt"

type ErrGreaterMaxDur struct {
	MaxDuration string
}

func NewGreaterMaxDur(max string) *ErrGreaterMaxDur {
	return &ErrGreaterMaxDur{MaxDuration: max}
}

func (e ErrGreaterMaxDur) Error() string {
	return fmt.Sprintf("end-start time should be less than %s, you can modify the configuration file ./config/strategy.toml 'max_duration'", e.MaxDuration)
}

type ErrLessMinDur struct {
	MinDuration string
}

func NewLessMinDur(min string) *ErrLessMinDur {
	return &ErrLessMinDur{MinDuration: min}
}

func (e ErrLessMinDur) Error() string {
	return fmt.Sprintf("end-start time should be greater than %s, you can modify the configuration file ./config/strategy.toml 'min_duration'", e.MinDuration)
}
