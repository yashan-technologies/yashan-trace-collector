package baseinfo

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/errdef"
	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/baseinfo/gopsutil"
	"ytc/internal/modules/ytc/collect/baseinfo/sar"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/execerutil"
	"ytc/utils/osutil"
	"ytc/utils/stringutil"
	"ytc/utils/userutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/fs"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"gopkg.in/ini.v1"
)

const (
	_tips_apt_base_host_load_status = "sudo apt install sysstat"
	_tips_yum_base_host_load_status = "sudo yum install sysstat"
	_tips_base_host_firewalld       = "you can run 'sudo ytcctl collect' or run 'ytcctl collect' with root"
)

const (
	KEY_YASDB_INI       = "yasdb.ini"
	KEY_YASDB_PARAMETER = "v$parameter"

	KEY_CURRENT = "current"
	KEY_HISTORY = "history"
)

const (
	_firewalld_inactive      = "inactive"
	_firewalld_active        = "active"
	_ubuntu_firewalld_active = "Status: active"
)

var (
	BaseInfoChineseName = map[string]string{
		datadef.BASE_YASDB_VERION:      "数据库版本",
		datadef.BASE_YASDB_PARAMETER:   "数据库配置",
		datadef.BASE_HOST_OS_INFO:      "操作系统信息",
		datadef.BASE_HOST_FIREWALLD:    "防火墙配置",
		datadef.BASE_HOST_CPU:          "CPU",
		datadef.BASE_HOST_DISK:         "磁盘",
		datadef.BASE_HOST_NETWORK:      "网络配置",
		datadef.BASE_HOST_MEMORY:       "内存",
		datadef.BASE_HOST_NETWORK_IO:   "网络流量",
		datadef.BASE_HOST_CPU_USAGE:    "CPU占用分析",
		datadef.BASE_HOST_DISK_IO:      "磁盘I/O",
		datadef.BASE_HOST_MEMORY_USAGE: "内存容量检查",
	}

	BaseInfoChildChineseName = map[string]string{
		KEY_YASDB_INI:       "数据库实例配置文件：yasdb.ini",
		KEY_YASDB_PARAMETER: "数据库实例参数视图：v$parameter",
		KEY_HISTORY:         "历史负载",
		KEY_CURRENT:         "当前负载",
	}
)

var ItemNameToWorkloadTypeMap = map[string]collecttypedef.WorkloadType{
	datadef.BASE_HOST_CPU_USAGE:    collecttypedef.WT_CPU,
	datadef.BASE_HOST_DISK_IO:      collecttypedef.WT_DISK,
	datadef.BASE_HOST_MEMORY_USAGE: collecttypedef.WT_MEMORY,
	datadef.BASE_HOST_NETWORK_IO:   collecttypedef.WT_NETWORK,
}

var WorkloadTypeToSarArgMap = map[collecttypedef.WorkloadType]string{
	collecttypedef.WT_CPU:     "-u",
	collecttypedef.WT_DISK:    "-d",
	collecttypedef.WT_MEMORY:  "-r",
	collecttypedef.WT_NETWORK: "-n DEV",
}

type checkFunc func() *ytccollectcommons.NoAccessRes

type DiskUsage struct {
	Device       string
	MountOptions string
	disk.UsageStat
}

type HostWorkResponse struct {
	Data     map[string]interface{}
	Errors   map[string]string
	DataType datadef.DataType
}

type BaseCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *datadef.YTCModule
	yasdbValidateErr error
	notConnectDB     bool
}

func NewBaseCollecter(collectParam *collecttypedef.CollectParam) *BaseCollecter {
	return &BaseCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &datadef.YTCModule{
			Module: collecttypedef.TYPE_BASE,
		},
	}
}

func (b *BaseCollecter) CheckAccess(yasdbValidate error) (noAccess []ytccollectcommons.NoAccessRes) {
	b.yasdbValidateErr = yasdbValidate
	noAccess = make([]ytccollectcommons.NoAccessRes, 0)
	funcMap := b.CheckFunc()
	for item, fn := range funcMap {
		noAccessRes := fn()
		if noAccessRes != nil {
			log.Module.Debugf("item [%s] check asscess desc: %s tips %s", item, noAccessRes.Description, noAccessRes.Tips)
			noAccess = append(noAccess, *noAccessRes)
		}
	}
	return
}

// [Interface Func]
func (b *BaseCollecter) CollectFunc(items []string) (res map[string]func() error) {
	res = make(map[string]func() error)
	itemFuncMap := b.itemFunc()
	for _, collectItem := range items {
		_, ok := itemFuncMap[collectItem]
		if !ok {
			log.Module.Errorf("get %s collect func err %s", collectItem)
			continue
		}
		res[collectItem] = itemFuncMap[collectItem]
	}
	return
}

// [Interface Func]
func (b *BaseCollecter) Type() string {
	return collecttypedef.TYPE_BASE
}

// [Interface Func]
func (b *BaseCollecter) CollectedItem(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range BaseInfoChineseName {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *BaseCollecter) getNotAccessItem(noAccess []ytccollectcommons.NoAccessRes) (res map[string]struct{}) {
	res = make(map[string]struct{})
	for _, noAccessRes := range noAccess {
		if noAccessRes.ForceCollect {
			continue
		}
		res[noAccessRes.ModuleItem] = struct{}{}
	}
	return
}

func (b *BaseCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.BASE_YASDB_VERION:      b.yasdbVersion,
		datadef.BASE_YASDB_PARAMETER:   b.yasdbParameter,
		datadef.BASE_HOST_OS_INFO:      b.hostOSInfo,
		datadef.BASE_HOST_FIREWALLD:    b.hostFirewalldStatus,
		datadef.BASE_HOST_CPU:          b.hostCPUInfo,
		datadef.BASE_HOST_DISK:         b.hostDiskInfo,
		datadef.BASE_HOST_NETWORK:      b.hostNetworkInfo,
		datadef.BASE_HOST_MEMORY:       b.hostMemoryInfo,
		datadef.BASE_HOST_NETWORK_IO:   b.hostNetworkIO,
		datadef.BASE_HOST_CPU_USAGE:    b.hostCPUUsage,
		datadef.BASE_HOST_DISK_IO:      b.hostDiskIO,
		datadef.BASE_HOST_MEMORY_USAGE: b.hostMemoryUsage,
	}
}

// [Interface Func]
func (b *BaseCollecter) Start(packageDir string) (err error) {
	return
}

// [Interface Func]
func (b *BaseCollecter) Finish() *datadef.YTCModule {
	return b.ModuleCollectRes
}

func (b *BaseCollecter) fillResult(data *datadef.YTCItem) {
	b.ModuleCollectRes.Set(data)
}

func (b *BaseCollecter) yasdbVersion() (err error) {
	yasdbVersionItem := datadef.YTCItem{Name: datadef.BASE_YASDB_VERION}
	defer b.fillResult(&yasdbVersionItem)

	log := log.Module.M(datadef.BASE_YASDB_VERION)
	yasdbBinPath := path.Join(b.YasdbHome, "bin", bashdef.CMD_YASDB)
	if !fs.IsFileExist(yasdbBinPath) {
		err = &errdef.ErrFileNotFound{Fname: yasdbBinPath}
		log.Errorf("failed to get yashandb version, err: %s", err.Error())
		yasdbVersionItem.Error = err.Error()
		// TODO: 补充description信息
		return
	}
	execer := execerutil.NewExecer(log)
	env := []string{fmt.Sprintf("%s=%s", yasqlutil.LIB_KEY, path.Join(b.YasdbHome, yasqlutil.LIB_PATH))}
	ret, stdout, stderr := execer.EnvExec(env, yasdbBinPath, "-V")
	if ret != 0 {
		err = fmt.Errorf("failed to get yasdb version, err: %s", stderr)
		log.Error(err)
		yasdbVersionItem.Error = stderr
		return
	}
	yasdbVersionItem.Details = strings.TrimSpace(stdout)
	return
}

func (b *BaseCollecter) yasdbParameter() (err error) {
	yasdbParameterItem := datadef.YTCItem{
		Name:     datadef.BASE_YASDB_PARAMETER,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&yasdbParameterItem)
	log := log.Module.M(datadef.BASE_YASDB_PARAMETER)

	// collect yasdb ini config
	if yasdbIni, err := b.getYasdbIni(); err != nil {
		yasdbParameterItem.Children[KEY_YASDB_INI] = datadef.YTCItem{Error: err.Error()}
		log.Errorf("failed to get yasdb.ini, err: %s", err.Error())
	} else {
		yasdbParameterItem.Children[KEY_YASDB_INI] = datadef.YTCItem{Details: yasdbIni}
	}

	// collect parameter from v$parameter
	if !b.notConnectDB {
		if pv, err := b.getYasdbParameter(); err != nil {
			yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Error: err.Error()}
			log.Errorf("failed to get yashandb parameter, err: %s", err.Error())
		} else {
			yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Details: pv}
		}
	} else {
		yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Error: "cannot connect to database"}
	}
	return
}

func (b *BaseCollecter) getYasdbIni() (res map[string]string, err error) {
	iniConfigPath := path.Join(b.YasdbData, "config", "yasdb.ini")
	res = make(map[string]string)
	if !fs.IsFileExist(iniConfigPath) {
		err = &errdef.ErrFileNotFound{Fname: iniConfigPath}
		return
	}
	yasdbConf, err := ini.Load(iniConfigPath)
	if err != nil {
		return
	}
	for _, section := range yasdbConf.Sections() {
		for _, key := range section.Keys() {
			res[key.Name()] = key.String()
		}
	}
	return
}

func (b *BaseCollecter) getYasdbParameter() (pv []*yasdb.VParameter, err error) {
	// collect parameter from v$parameter
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	return yasdb.QueryAllParameter(tx)
}

func (b *BaseCollecter) hostOSInfo() (err error) {
	hostBaseInfoItem := datadef.YTCItem{Name: datadef.BASE_HOST_OS_INFO}
	defer b.fillResult(&hostBaseInfoItem)

	log := log.Module.M(datadef.BASE_HOST_OS_INFO)
	hostInfo, err := host.Info()
	if err != nil {
		log.Errorf("failed to get host os info, err: %s", err.Error())
		hostBaseInfoItem.Error = err.Error()
		return
	}
	hostBaseInfoItem.Details = hostInfo
	return
}

func (b *BaseCollecter) hostFirewalldStatus() (err error) {
	hostFirewallStatus := datadef.YTCItem{Name: datadef.BASE_HOST_FIREWALLD}
	defer b.fillResult(&hostFirewallStatus)

	log := log.Module.M(datadef.BASE_HOST_FIREWALLD)
	osRelease, err := osutil.GetOsRelease()
	if err != nil {
		log.Errorf("failed to get host os release info, err: %s", err.Error())
		hostFirewallStatus.Error = err.Error()
		return
	}
	execer := execerutil.NewExecer(log)
	// ubuntu
	if osRelease.Id == osutil.UBUNTU_ID {
		if !userutil.IsCurrentUserRoot() {
			hostFirewallStatus.Error = "checking ubuntu firewall status requires sudo or root"
			hostFirewallStatus.Description = "查看Ubuntu系统防火墙状态需要root权限"
			return
		}
		_, stdout, _ := execer.Exec(bashdef.CMD_BASH, "-c", fmt.Sprintf("%s status", bashdef.CMD_UFW))
		hostFirewallStatus.Details = strings.Contains(stdout, _ubuntu_firewalld_active)
		return
	}
	// other os
	_, stdout, _ := execer.Exec(bashdef.CMD_BASH, "-c", fmt.Sprintf("%s is-active firewalld", bashdef.CMD_SYSTEMCTL))
	hostFirewallStatus.Details = strings.Contains(stdout, _firewalld_active) && !strings.Contains(stdout, _firewalld_inactive)
	return
}

func (b *BaseCollecter) hostCPUInfo() (err error) {
	hostCpuInfo := datadef.YTCItem{Name: datadef.BASE_HOST_CPU}
	defer b.fillResult(&hostCpuInfo)

	log := log.Module.M(datadef.BASE_HOST_CPU)
	cpuInfo, err := cpu.Info()
	if err != nil {
		log.Errorf("failed to get host cpu info, err: %s", err.Error())
		hostCpuInfo.Error = err.Error()
		return
	}
	hostCpuInfo.Details = cpuInfo
	return
}

func (b *BaseCollecter) hostDiskInfo() (err error) {
	hostDiskInfo := datadef.YTCItem{Name: datadef.BASE_HOST_DISK}
	defer b.fillResult(&hostDiskInfo)

	log := log.Module.M(datadef.BASE_HOST_DISK)
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Errorf("failed to get host disk info, err: %s", err.Error())
		hostDiskInfo.Error = err.Error()
		return
	}
	var usages []DiskUsage
	for _, partition := range partitions {
		var usageStat *disk.UsageStat
		usageStat, err = disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Errorf("failed to get disk usage info, err: %s", err.Error())
			hostDiskInfo.Error = err.Error()
			return
		}
		usage := DiskUsage{
			Device:       partition.Device,
			MountOptions: partition.Opts,
			UsageStat:    *usageStat,
		}
		usages = append(usages, usage)
	}
	hostDiskInfo.Details = usages
	return
}

func (b *BaseCollecter) hostNetworkInfo() (err error) {
	hostNetInfo := datadef.YTCItem{Name: datadef.BASE_HOST_NETWORK}
	defer b.fillResult(&hostNetInfo)

	log := log.Module.M(datadef.BASE_HOST_NETWORK)
	netInfo, err := net.Interfaces()
	if err != nil {
		log.Errorf("failed to get host network info, err: %s", err.Error())
		hostNetInfo.Error = err.Error()
		return
	}
	hostNetInfo.Details = netInfo
	return
}

func (b *BaseCollecter) hostMemoryInfo() (err error) {
	hostMemoryInfo := datadef.YTCItem{Name: datadef.BASE_HOST_MEMORY}
	defer b.fillResult(&hostMemoryInfo)

	log := log.Module.M(datadef.BASE_HOST_MEMORY)
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("failed to get host memory info: %s", err.Error())
		hostMemoryInfo.Error = err.Error()
		return
	}
	hostMemoryInfo.Details = memInfo
	return
}

func (b *BaseCollecter) hostCPUUsage() (err error) {
	hostCPUUsage := datadef.YTCItem{
		Name:     datadef.BASE_HOST_CPU_USAGE,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostCPUUsage)

	log := log.Module.M(datadef.BASE_HOST_CPU_USAGE)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_CPU_USAGE)
	if err != nil {
		log.Error("failed to get host cpu usage info, err: %s", err.Error())
		hostCPUUsage.Error = err.Error()
		return
	}
	hostCPUUsage.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostCPUUsage.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}

func (b *BaseCollecter) hostNetworkIO() (err error) {
	hostNetworkIO := datadef.YTCItem{
		Name:     datadef.BASE_HOST_NETWORK_IO,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostNetworkIO)

	log := log.Module.M(datadef.BASE_HOST_NETWORK_IO)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_NETWORK_IO)
	if err != nil {
		log.Errorf("failed to get host network IO info, err: %s", err.Error())
		hostNetworkIO.Error = err.Error()
		return
	}
	hostNetworkIO.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostNetworkIO.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}

func (b *BaseCollecter) hostDiskIO() (err error) {
	hostDiskIO := datadef.YTCItem{
		Name:     datadef.BASE_HOST_DISK_IO,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostDiskIO)

	log := log.Module.M(datadef.BASE_HOST_DISK_IO)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_DISK_IO)
	if err != nil {
		log.Error("failed to get host disk IO info, err: %s", err.Error())
		hostDiskIO.Error = err.Error()
		return
	}
	hostDiskIO.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostDiskIO.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}

func (b *BaseCollecter) hostMemoryUsage() (err error) {
	hostMemoryUsage := datadef.YTCItem{
		Name:     datadef.BASE_HOST_MEMORY_USAGE,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&hostMemoryUsage)

	log := log.Module.M(datadef.BASE_HOST_MEMORY_USAGE)
	resp, err := b.hostWorkload(log, datadef.BASE_HOST_MEMORY_USAGE)
	if err != nil {
		log.Errorf("failed to gert host memory usage info, err: %s", err.Error())
		hostMemoryUsage.Error = err.Error()
		return
	}
	hostMemoryUsage.Children[KEY_HISTORY] = datadef.YTCItem{
		Error:    resp.Errors[KEY_HISTORY],
		Details:  resp.Data[KEY_HISTORY],
		DataType: resp.DataType,
	}
	hostMemoryUsage.Children[KEY_CURRENT] = datadef.YTCItem{
		Error:    resp.Errors[KEY_CURRENT],
		Details:  resp.Data[KEY_CURRENT],
		DataType: resp.DataType,
	}
	return
}

func (b *BaseCollecter) hostWorkload(log yaslog.YasLog, itemName string) (resp HostWorkResponse, err error) {
	details := map[string]interface{}{}
	hasSar := b.CheckSarAccess() == nil
	resp.DataType = datadef.DATATYPE_GOPSUTIL
	resp.Errors = make(map[string]string)

	// collect historyworkload
	if hasSar {
		resp.DataType = datadef.DATATYPE_SAR
		if historyNetworkWorkload, e := b.hostHistoryWorkload(log, itemName, b.StartTime, b.EndTime); e != nil {
			err = fmt.Errorf("failed to collect history %s, err: %s", itemName, e.Error())
			resp.Errors[KEY_HISTORY] = err.Error()
			log.Error(err)
		} else {
			details[KEY_HISTORY] = historyNetworkWorkload
		}
	} else {
		err = fmt.Errorf("cannot find command '%s'", bashdef.CMD_SAR)
		resp.Errors[KEY_HISTORY] = err.Error()
		log.Error(err)
	}

	// collect current workload
	if currentNetworkWorkload, e := b.hostCurrentWorkload(log, itemName, hasSar); e != nil {
		err = fmt.Errorf("failed to collect current %s, err: %s", itemName, e.Error())
		resp.Errors[KEY_CURRENT] = err.Error()
		log.Error(err)
	} else {
		details[KEY_CURRENT] = currentNetworkWorkload
	}
	resp.Data = details
	return
}

func (b *BaseCollecter) hostHistoryWorkload(log yaslog.YasLog, itemName string, start, end time.Time) (resp collecttypedef.WorkloadOutput, err error) {
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err = fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return
	}
	sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
	if !ok {
		err = fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
		log.Error(err)
		return
	}
	// collect
	sar := sar.NewSar(log)
	strategyConf := confdef.GetStrategyConf()
	sarDir := strategyConf.Collect.SarDir
	if stringutil.IsEmpty(sarDir) {
		sarDir = sar.GetSarDir()
	}
	sarOutput := make(collecttypedef.WorkloadOutput)
	args := b.genHistoryWorkloadArgs(start, end, sarDir)
	for _, arg := range args {
		output, e := sar.Collect(workloadType, sarArg, arg)
		if e != nil {
			log.Error(e)
			continue
		}
		for timestamp, output := range output {
			sarOutput[timestamp] = output
		}
	}
	resp = sarOutput
	return
}

// TODO:这里的时区要需要处理
func (b *BaseCollecter) genHistoryWorkloadArgs(start, end time.Time, sarDir string) (args []string) {
	// get data between start and end
	var dates []time.Time
	begin := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	for date := begin; !date.After(end); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date)
	}
	for i, date := range dates {
		var startArg, endArg, fileArg string
		if i == 0 && !date.Equal(start) { // the frist
			startArg = fmt.Sprintf("-s %s", start.Format(timedef.TIME_FORMAT_TIME))
		}
		if i == len(dates)-1 { // the last one
			if date.Equal(end) { // skip
				continue
			}
			endArg = fmt.Sprintf("-e %s", end.Format(timedef.TIME_FORMAT_TIME))
		}
		fileArg = fmt.Sprintf("-f %s", path.Join(sarDir, fmt.Sprintf("sa%s", date.Format(timedef.TIME_FORMAT_DAY))))
		args = append(args, fmt.Sprintf("%s %s %s", fileArg, startArg, endArg))
	}
	return
}

func (b *BaseCollecter) hostCurrentWorkload(log yaslog.YasLog, itemName string, hasSar bool) (resp collecttypedef.WorkloadOutput, err error) {
	// global conf
	strategyConf := confdef.GetStrategyConf()
	scrapeInterval, scrapeTimes := strategyConf.Collect.ScrapeInterval, strategyConf.Collect.ScrapeTimes
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err = fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return
	}
	if hasSar { // use sar to collect first
		sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
		if !ok {
			err = fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
			log.Error(err)
			return
		}
		sar := sar.NewSar(log)
		return sar.Collect(workloadType, sarArg, strconv.Itoa(scrapeInterval), strconv.Itoa(scrapeTimes))
	}
	// use gopsutil to calculate by ourself
	return gopsutil.Collect(workloadType, scrapeInterval, scrapeTimes)
}
