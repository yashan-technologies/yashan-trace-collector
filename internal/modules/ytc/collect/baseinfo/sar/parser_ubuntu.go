package sar

import (
	"ytc/defs/collecttypedef"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
)

type UbuntuParser struct {
	base *baseParser
}

func NewUbuntuParser(yaslog yaslog.YasLog) *UbuntuParser {
	return &UbuntuParser{
		base: NewBaseParser(yaslog),
	}
}

func (u *UbuntuParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	return u.base.GetParserFunc(t)
}

func (u *UbuntuParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return u.base.ParseCpu(m, values)
}

func (u *UbuntuParser) IsCpuTitle(line string) bool {
	return u.base.IsCpuTitle(line)
}

func (u *UbuntuParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return u.base.ParseNetwork(m, values)
}

func (u *UbuntuParser) IsNetworkTitle(line string) bool {
	return u.base.IsNetworkTitle(line)
}

func (u *UbuntuParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return u.base.ParseDisk(m, values)
}

func (u *UbuntuParser) IsDiskTitle(line string) bool {
	return u.base.IsDiskTitle(line)
}

func (u *UbuntuParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	return u.base.ParseMemory(m, values)
}

func (u *UbuntuParser) IsMemoryTitle(line string) bool {
	return u.base.IsMemoryTitle(line)
}

func (u *UbuntuParser) GetSarDir() string {
	defaultConfigPath := "/etc/sysstat/sysstat"
	defaultFilePath := "/var/log/sysstat"
	currentFilePath := u.base.getSarDirFromConfig(defaultConfigPath)
	if stringutil.IsEmpty(currentFilePath) {
		return defaultFilePath
	}
	return currentFilePath
}
