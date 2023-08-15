package sar

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"ytc/defs/collecttypedef"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/fs"
)

type SarParser interface {
	GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc)
	ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem
	IsCpuTitle(line string) bool
	ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem
	IsNetworkTitle(line string) bool
	ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem
	IsMemoryTitle(line string) bool
	ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem
	IsDiskTitle(line string) bool
	GetSarDir() string
}

// use centos as base parser
type baseParser struct {
	log yaslog.YasLog
}

func NewBaseParser(yaslog yaslog.YasLog) *baseParser {
	return &baseParser{
		log: yaslog,
	}
}

func (b *baseParser) GetParserFunc(t collecttypedef.WorkloadType) (SarParseFunc, SarCheckTitleFunc) {
	switch t {
	case collecttypedef.WT_DISK:
		return b.ParseDisk, b.IsDiskTitle
	case collecttypedef.WT_MEMORY:
		return b.ParseMemory, b.IsMemoryTitle
	case collecttypedef.WT_NETWORK:
		return b.ParseNetwork, b.IsNetworkTitle
	case collecttypedef.WT_CPU:
		return b.ParseCpu, b.IsCpuTitle
	default:
		return nil, nil
	}
}

func (b *baseParser) ParseCpu(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	if len(values) < 9 { // not enough data, skip
		b.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	cpuUsage := CpuUsage{}
	cpuUsage.CPU = values[2]
	cpuUsage.User, _ = strconv.ParseFloat(values[3], 64)
	cpuUsage.Nice, _ = strconv.ParseFloat(values[4], 64)
	cpuUsage.System, _ = strconv.ParseFloat(values[5], 64)
	cpuUsage.IOWait, _ = strconv.ParseFloat(values[6], 64)
	cpuUsage.Steal, _ = strconv.ParseFloat(values[7], 64)
	cpuUsage.Idle, _ = strconv.ParseFloat(values[8], 64)
	m[cpuUsage.CPU] = cpuUsage
	return m
}

func (b *baseParser) IsCpuTitle(line string) bool {
	return strings.Contains(line, "CPU")
}

func (b *baseParser) ParseNetwork(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	if len(values) < 10 { // not enough data, skip
		b.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	networkIO := NetworkIO{}
	networkIO.Iface = values[2] // fill data
	networkIO.Rxpck, _ = strconv.ParseFloat(values[3], 64)
	networkIO.Txpck, _ = strconv.ParseFloat(values[4], 64)
	networkIO.RxkB, _ = strconv.ParseFloat(values[5], 64)
	networkIO.TxkB, _ = strconv.ParseFloat(values[6], 64)
	networkIO.Rxcmp, _ = strconv.ParseFloat(values[7], 64)
	networkIO.Txcmp, _ = strconv.ParseFloat(values[8], 64)
	networkIO.Rxmcst, _ = strconv.ParseFloat(values[9], 64)
	m[networkIO.Iface] = networkIO
	return m
}

func (b *baseParser) IsNetworkTitle(line string) bool {
	return strings.Contains(line, "IFACE")
}

func (b *baseParser) ParseMemory(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	if len(values) < 12 {
		b.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	memoryUsage := MemoryUsage{}
	memoryUsage.KBMemFree, _ = strconv.ParseInt(values[2], 10, 64)
	memoryUsage.KBmemUsed, _ = strconv.ParseInt(values[3], 10, 64)
	memoryUsage.MemUsed, _ = strconv.ParseFloat(values[4], 64)
	memoryUsage.KBBuffers, _ = strconv.ParseInt(values[5], 10, 64)
	memoryUsage.KBCached, _ = strconv.ParseInt(values[6], 10, 64)
	memoryUsage.KBCommit, _ = strconv.ParseInt(values[7], 10, 64)
	memoryUsage.Commit, _ = strconv.ParseFloat(values[8], 64)
	memoryUsage.KBActive, _ = strconv.ParseInt(values[9], 10, 64)
	memoryUsage.KBInact, _ = strconv.ParseInt(values[10], 10, 64)
	memoryUsage.KBDirty, _ = strconv.ParseInt(values[11], 10, 64)
	return m
}

func (b *baseParser) IsMemoryTitle(line string) bool {
	return strings.Contains(line, "kbmemfree")
}

func (b *baseParser) ParseDisk(m collecttypedef.WorkloadItem, values []string) collecttypedef.WorkloadItem {
	if len(values) < 11 {
		b.log.Warnf("not enough data, skip line: %s", strings.Join(values, stringutil.STR_BLANK_SPACE))
		return m
	}
	diskIO := DiskIO{}
	diskIO.Dev = values[2]
	diskIO.Tps, _ = strconv.ParseFloat(values[3], 64)
	diskIO.RdSec, _ = strconv.ParseFloat(values[4], 64)
	diskIO.WrSec, _ = strconv.ParseFloat(values[5], 64)
	diskIO.AvgrqSz, _ = strconv.ParseFloat(values[6], 64)
	diskIO.AvgquSz, _ = strconv.ParseFloat(values[7], 64)
	diskIO.Await, _ = strconv.ParseFloat(values[8], 64)
	diskIO.Svctm, _ = strconv.ParseFloat(values[9], 64)
	diskIO.Util, _ = strconv.ParseFloat(values[10], 64)
	m[diskIO.Dev] = diskIO
	return m
}

func (b *baseParser) IsDiskTitle(line string) bool {
	return strings.Contains(line, "DEV") && strings.Contains(line, "tps")
}

func (b *baseParser) GetSarDir() string {
	defaultConfigPath := "/etc/sysconfig/sysstat"
	defaultSarDir := "/var/log/sa"
	currentFilePath := b.getSarDirFromConfig(defaultConfigPath)
	if stringutil.IsEmpty(currentFilePath) {
		return defaultSarDir
	}
	return currentFilePath
}

func (b *baseParser) getSarDirFromConfig(configPath string) string {
	if !fs.IsFileExist(configPath) {
		return ""
	}
	configMap := make(map[string]string)
	file, err := os.Open(configPath)
	if err != nil {
		b.log.Error(err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") { // ignore
			continue
		}
		re := regexp.MustCompile(`^([^=]+)=(.*)$`) // key=value
		match := re.FindStringSubmatch(line)
		if len(match) == 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			configMap[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		b.log.Error(err)
		return ""
	}
	return configMap["SAR_DIR"]
}
