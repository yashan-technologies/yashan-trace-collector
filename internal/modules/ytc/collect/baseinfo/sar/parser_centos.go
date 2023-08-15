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

func (c *CentosParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	return c.base.GetParserFunc(t)
}

func (c *CentosParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return c.base.ParseCpu(m, values)
}

func (c *CentosParser) IsCpuTitle(line string) bool {
	return c.base.IsCpuTitle(line)
}

func (c *CentosParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return c.base.ParseNetwork(m, values)

}

func (c *CentosParser) IsNetworkTitle(line string) bool {
	return c.base.IsNetworkTitle(line)
}

func (c *CentosParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return c.base.ParseDisk(m, values)
}

func (c *CentosParser) IsDiskTitle(line string) bool {
	return c.base.IsDiskTitle(line)
}

func (c *CentosParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return c.base.ParseMemory(m, values)
}

func (c *CentosParser) IsMemoryTitle(line string) bool {
	return c.base.IsMemoryTitle(line)
}

func (c *CentosParser) GetSarDir() string {
	return c.base.GetSarDir()
}
