package errdef

type ErrYasdbProcessNotFound struct {
}

func NewErrYasdbProcessNotFound() *ErrYasdbProcessNotFound {
	return &ErrYasdbProcessNotFound{}
}

func (e *ErrYasdbProcessNotFound) Error() string {
	return "yasdb process not found"
}
