// package baseinforeporter is used to generate the base info reports
package baseinforeporter

import (
	"fmt"

	"ytc/internal/modules/ytc/collect/baseinfo"
	"ytc/internal/modules/ytc/collect/commons/datadef"
)

func validateWorkLoadItem(item datadef.YTCItem) (historyItem, currentItem datadef.YTCItem, err error) {
	if len(item.Children) == 0 {
		err = fmt.Errorf("invalid data, children of %s unfound", item.Name)
		return
	}
	historyItem, ok := item.Children[baseinfo.KEY_HISTORY]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", baseinfo.KEY_HISTORY, item.Children)
		return
	}
	currentItem, ok = item.Children[baseinfo.KEY_CURRENT]
	if !ok {
		err = fmt.Errorf("invalid data, %s unfound in %v", baseinfo.KEY_CURRENT, item.Children)
		return
	}
	return
}
