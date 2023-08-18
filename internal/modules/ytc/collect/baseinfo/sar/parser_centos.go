package sar

import (
	"ytc/defs/collecttypedef"

	"git.yasdb.com/go/yaslog"
)

type CentosParser struct {
	base *baseParser
}

func NewCentosParser(yaslog yaslog.YasLog) *CentosParser {
	return &CentosParser{
		base: NewBaseParser(yaslog),
	}
}

// [Interface Func]
func (c *CentosParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	return c.base.GetParserFunc(t)
}

// [Interface Func]
func (c *CentosParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -u
	return c.base.ParseCpu(m, values)
}

// [Interface Func]
func (c *CentosParser) IsCpuTitle(line string) bool {
	return c.base.IsCpuTitle(line)
}

// [Interface Func]
func (c *CentosParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -n DEV
	return c.base.ParseNetwork(m, values)

}

// [Interface Func]
func (c *CentosParser) IsNetworkTitle(line string) bool {
	return c.base.IsNetworkTitle(line)
}

// [Interface Func]
func (c *CentosParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -d
	return c.base.ParseDisk(m, values)
}

// [Interface Func]
func (c *CentosParser) IsDiskTitle(line string) bool {
	return c.base.IsDiskTitle(line)
}

// [Interface Func]
func (c *CentosParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// commadn: sar -r
	return c.base.ParseMemory(m, values)
}

// [Interface Func]
func (c *CentosParser) IsMemoryTitle(line string) bool {
	return c.base.IsMemoryTitle(line)
}

// [Interface Func]
func (c *CentosParser) GetSarDir() string {
	return c.base.GetSarDir()
}
