package errdef

import "fmt"

type FormItemUnfound struct {
	ItemName string
}

func NewFormItemUnfound(itemName string) *FormItemUnfound {
	return &FormItemUnfound{
		ItemName: itemName,
	}
}

func (e *FormItemUnfound) Error() string {
	return fmt.Sprintf("form item %s unfound", e.ItemName)
}

type ItemEmpty struct {
	ItemName string
}

func NewItemEmpty(itemName string) *ItemEmpty {
	return &ItemEmpty{
		ItemName: itemName,
	}
}

func (e *ItemEmpty) Error() string {
	return fmt.Sprintf("%s is empty", e.ItemName)
}
