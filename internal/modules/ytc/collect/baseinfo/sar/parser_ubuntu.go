package sar

import (
	"strconv"
	"strings"

	"ytc/defs/collecttypedef"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
)

// title: dd:mm:ss AM DEV       tps     rkB/s     wkB/s   areq-sz    aqu-sz     await     svctm     %util
const (
	_ubuntu_disk_date_index OutputIndex = iota
	_ubuntu_disk_period_index
	_ubuntu_disk_dev_index
	_ubuntu_disk_tps_index
	_ubuntu_disk_rkb_index
	_ubuntu_disk_wkb_index
	_ubuntu_disk_areqsz_index
	_ubuntu_disk_aqusz_index
	_ubuntu_disk_await_index
	_ubuntu_disk_svctm_index
	_ubuntu_disk_util_index
	_ubuntu_disk_length
)

// dd:mm:ss AM kbmemfree   kbavail kbmemused  %memused kbbuffers  kbcached  kbcommit   %commit  kbactive   kbinact   kbdirty
const (
	_ubuntu_memory_date_index OutputIndex = iota
	_ubuntu_memory_period_index
	_ubuntu_memory_kbmemfree_index
	_ubuntu_memory_kbavail_index
	_ubuntu_memory_kbmemused_index
	_ubuntu_memory_memused_index
	_ubuntu_memory_kbbuffers_index
	_ubuntu_memory_kbcached_index
	_ubuntu_memory_kbcommit_index
	_ubuntu_memory_commit_index
	_ubuntu_memory_kbactive_index
	_ubuntu_memory_kbinact_index
	_ubuntu_memory_kbdirty_index
	_ubuntu_memory_length
)

type UbuntuParser struct {
	base *baseParser
}

func NewUbuntuParser(yaslog yaslog.YasLog) *UbuntuParser {
	return &UbuntuParser{
		base: NewBaseParser(yaslog),
	}
}

// [Interface Func]
func (u *UbuntuParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	switch t {
	case collecttypedef.WT_DISK:
		return u.ParseDisk, u.IsDiskTitle
	case collecttypedef.WT_MEMORY:
		return u.ParseMemory, u.IsMemoryTitle
	case collecttypedef.WT_NETWORK:
		return u.ParseNetwork, u.IsNetworkTitle
	case collecttypedef.WT_CPU:
		return u.ParseCpu, u.IsCpuTitle
	default:
		return nil, nil
	}
}

// [Interface Func]
func (u *UbuntuParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -u
	return u.base.ParseCpu(m, values)
}

// [Interface Func]
func (u *UbuntuParser) IsCpuTitle(line string) bool {
	return u.base.IsCpuTitle(line)
}

// [Interface Func]
func (u *UbuntuParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -n DEV
	return u.base.ParseNetwork(m, values)
}

// [Interface Func]
func (u *UbuntuParser) IsNetworkTitle(line string) bool {
	return u.base.IsNetworkTitle(line)
}

// [Interface Func]
func (u *UbuntuParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -d
	if len(values) < int(_ubuntu_disk_length) {
		u.base.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	diskIO := DiskIO{}
	var err error
	diskIO.Dev = values[_ubuntu_disk_dev_index]
	if diskIO.Tps, err = strconv.ParseFloat(values[_ubuntu_disk_tps_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.RKBSec, err = strconv.ParseFloat(values[_ubuntu_disk_rkb_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.WKBSec, err = strconv.ParseFloat(values[_ubuntu_disk_wkb_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.AvgrqSz, err = strconv.ParseFloat(values[_ubuntu_disk_areqsz_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.AvgquSz, err = strconv.ParseFloat(values[_ubuntu_disk_aqusz_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.Await, err = strconv.ParseFloat(values[_ubuntu_disk_await_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.Svctm, err = strconv.ParseFloat(values[_ubuntu_disk_svctm_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if diskIO.Util, err = strconv.ParseFloat(values[_ubuntu_disk_util_index], 64); err != nil {
		u.base.log.Error(err)
	}
	m[diskIO.Dev] = diskIO
	return m
}

// [Interface Func]
func (u *UbuntuParser) IsDiskTitle(line string) bool {
	return u.base.IsDiskTitle(line)
}

// [Interface Func]
func (u *UbuntuParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	// command: sar -r
	if len(values) < int(_ubuntu_memory_length) {
		u.base.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	memoryUsage := MemoryUsage{}
	var err error
	if memoryUsage.KBMemFree, err = strconv.ParseInt(values[_ubuntu_memory_kbmemfree_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBAvail, err = strconv.ParseInt(values[_ubuntu_memory_kbavail_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBmemUsed, err = strconv.ParseInt(values[_ubuntu_memory_kbmemused_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.MemUsed, err = strconv.ParseFloat(values[_ubuntu_memory_memused_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBBuffers, err = strconv.ParseInt(values[_ubuntu_memory_kbbuffers_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBCached, err = strconv.ParseInt(values[_ubuntu_memory_kbcached_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBCommit, err = strconv.ParseInt(values[_ubuntu_memory_kbcommit_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.Commit, err = strconv.ParseFloat(values[_ubuntu_memory_commit_index], 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBActive, err = strconv.ParseInt(values[_ubuntu_memory_kbactive_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBInact, err = strconv.ParseInt(values[_ubuntu_memory_kbinact_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	if memoryUsage.KBDirty, err = strconv.ParseInt(values[_ubuntu_memory_kbdirty_index], 10, 64); err != nil {
		u.base.log.Error(err)
	}
	m[memoryUsageKey] = u.base.calculateRealMemUsed(memoryUsage)
	return m
}

// [Interface Func]
func (u *UbuntuParser) IsMemoryTitle(line string) bool {
	return u.base.IsMemoryTitle(line)
}

// [Interface Func]
func (u *UbuntuParser) GetSarDir() string {
	defaultConfigPath := "/etc/sysstat/sysstat"
	defaultFilePath := "/var/log/sysstat"
	currentFilePath := u.base.getSarDirFromConfig(defaultConfigPath)
	if stringutil.IsEmpty(currentFilePath) {
		return defaultFilePath
	}
	return currentFilePath
}
