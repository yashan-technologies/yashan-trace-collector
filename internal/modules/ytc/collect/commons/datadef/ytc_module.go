package datadef

import "sync"

const (
	DATATYPE_SAR      DataType = "sar"
	DATATYPE_GOPSUTIL DataType = "gopstuil"
)

type DataType string

type YTCItem struct {
	Name        string             `json:"-"`                     // 收集项名称
	Error       string             `json:"error,omitempty"`       // 原始报错信息
	Description string             `json:"description,omitempty"` // 失败原因描述
	Details     interface{}        `json:"details,omitempty"`     // 每个收集项包含的数据
	DataType    DataType           `json:"datatype,omitempty"`    // 数据类型，在Details可能使用多种数据时使用
	Children    map[string]YTCItem `json:"children,omitempty"`
}

type YTCModule struct {
	Module    string `json:"-"`
	mtx       sync.RWMutex
	items     map[string]*YTCItem
	JSONItems map[string]*YTCItem `json:"items"`
}

func (c *YTCModule) Set(item *YTCItem) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.items == nil {
		c.items = make(map[string]*YTCItem)
	}
	c.items[item.Name] = item
}

func (c *YTCModule) FillJSONItems() {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	c.JSONItems = make(map[string]*YTCItem)
	for k, v := range c.items {
		tmp := *v
		c.JSONItems[k] = &tmp
	}
}

func (c *YTCModule) Items() map[string]*YTCItem {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	items := make(map[string]*YTCItem)
	for k, v := range c.items {
		tmp := *v
		items[k] = &tmp
	}
	return items
}
