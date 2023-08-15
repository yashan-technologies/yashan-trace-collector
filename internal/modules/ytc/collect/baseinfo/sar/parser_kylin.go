package sar

import (
	"ytc/defs/collecttypedef"

	"git.yasdb.com/go/yaslog"
)

type KylinParser struct {
	base *baseParser
}

func NewKylinParser(yaslog yaslog.YasLog) *KylinParser {
	return &KylinParser{
		base: NewBaseParser(yaslog),
	}
}
func (k *KylinParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	return k.base.GetParserFunc(t)
}

func (k *KylinParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return k.base.ParseCpu(m, values)
}

func (k *KylinParser) IsCpuTitle(line string) bool {
	return k.base.IsCpuTitle(line)
}

func (k *KylinParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return k.base.ParseNetwork(m, values)
}

func (k *KylinParser) IsNetworkTitle(line string) bool {
	return k.base.IsNetworkTitle(line)
}

func (k *KylinParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return k.base.ParseDisk(m, values)
}

func (k *KylinParser) IsDiskTitle(line string) bool {
	return k.base.IsDiskTitle(line)
}

func (k *KylinParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return k.base.ParseMemory(m, values)
}

func (k *KylinParser) IsMemoryTitle(line string) bool {
	return k.base.IsMemoryTitle(line)
}

func (k *KylinParser) GetSarDir() string {
	return k.base.GetSarDir()
}
