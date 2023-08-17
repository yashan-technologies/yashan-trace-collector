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
	"ytc/internal/modules/ytc/collect/data"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/osutil"
	"ytc/utils/stringutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yaslog"
	"git.yasdb.com/go/yasutil/execer"
	"git.yasdb.com/go/yasutil/fs"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"gopkg.in/ini.v1"
)

type checkFunc func() *data.NoAccessRes

const (
	_tips_apt_base_host_load_status = "sudo apt install sysstat"
	_tips_yum_base_host_load_status = "sudo yum install sysstat"
	_tips_base_host_firewalld       = "you can run 'sudo ytcctl collect' or run 'ytcctl collect' with root"
)

const (
	_yasdb_ini_key       = "yasdb.ini"
	_yasdb_parameter_key = "v$parameter"

	_current_key = "current"
	_history_key = "history"

	_firewalld_inactive      = "inactive"
	_firewalld_active        = "active"
	_ubuntu_firewalld_active = "Status: active"
)

var (
	base_default_ch_map = map[string]string{
		data.BASE_YASDB_VERION:      "数据库版本",
		data.BASE_YASDB_PARAMTER:    "数据库服务端配置参数",
		data.BASE_HOST_OS_INFO:      "操作系统信息",
		data.BASE_HOST_FIREWALLD:    "防火墙状态",
		data.BASE_HOST_CPU:          "CPU",
		data.BASE_HOST_DISK:         "磁盘",
		data.BASE_HOST_Network:      "网卡",
		data.BASE_HOST_Memery:       "内存",
		data.BASE_HOST_NETWORK_IO:   "网络流量",
		data.BASE_HOST_CPU_USAGE:    "CPU占用分析",
		data.BASE_HOST_DISK_IO:      "磁盘IO",
		data.BASE_HOST_MEMORY_USAGE: "内存容量检查",
	}
)

var ItemNameToWorkloadTypeMap = map[string]collecttypedef.WorkloadType{
	data.BASE_HOST_CPU_USAGE:    collecttypedef.WT_CPU,
	data.BASE_HOST_DISK_IO:      collecttypedef.WT_DISK,
	data.BASE_HOST_MEMORY_USAGE: collecttypedef.WT_MEMORY,
	data.BASE_HOST_NETWORK_IO:   collecttypedef.WT_NETWORK,
}

var WorkloadTypeToSarArgMap = map[collecttypedef.WorkloadType]string{
	collecttypedef.WT_CPU:     "-u",
	collecttypedef.WT_DISK:    "-d",
	collecttypedef.WT_MEMORY:  "-r",
	collecttypedef.WT_NETWORK: "-n DEV",
}

type BaseCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *data.YtcModule
	yasdbValidateErr error
	notConnectDB     bool
}

// data from linux command 'ps -aux'
type PsResult struct {
	User       string
	Pid        string
	CpuPercent float64
	MemPercent float64
	Vsz        int64
	Kss        int64
	Tty        string
	Stat       string
	Start      string
	Time       string
	Command    string
}

func NewBaseCollecter(collectParam *collecttypedef.CollectParam) *BaseCollecter {
	return &BaseCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &data.YtcModule{
			Module: collecttypedef.TYPE_BASE,
		},
	}
}

func (b *BaseCollecter) CheckAccess(yasdbValidate error) (noAccess []data.NoAccessRes) {
	b.yasdbValidateErr = yasdbValidate
	noAccess = make([]data.NoAccessRes, 0)
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

func (b *BaseCollecter) Type() string {
	return collecttypedef.TYPE_BASE
}

func (b *BaseCollecter) CollectedItem(noAccess []data.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range base_default_ch_map {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *BaseCollecter) getNotAccessItem(noAccess []data.NoAccessRes) (res map[string]struct{}) {
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
		data.BASE_YASDB_VERION:      b.yasdbVersion,
		data.BASE_YASDB_PARAMTER:    b.yasdbPrameter,
		data.BASE_HOST_OS_INFO:      b.hostOSInfo,
		data.BASE_HOST_FIREWALLD:    b.hostFirewalldStatus,
		data.BASE_HOST_CPU:          b.hostCPUInfo,
		data.BASE_HOST_DISK:         b.hostDiskInfo,
		data.BASE_HOST_Network:      b.hostNetworkInfo,
		data.BASE_HOST_Memery:       b.hostMemoryInfo,
		data.BASE_HOST_NETWORK_IO:   b.hostNetworkIO,
		data.BASE_HOST_CPU_USAGE:    b.hostCPUUsage,
		data.BASE_HOST_DISK_IO:      b.hostDiskIO,
		data.BASE_HOST_MEMORY_USAGE: b.hostMemoryUsage,
	}
}

func (b *BaseCollecter) Start(packageDir string) error {
	return nil
}

func (b *BaseCollecter) Finish() *data.YtcModule {
	return b.ModuleCollectRes
}

func (b *BaseCollecter) fillResult(data *data.YtcItem) {
	b.ModuleCollectRes.Lock()
	defer b.ModuleCollectRes.Unlock()
	b.ModuleCollectRes.Items = append(b.ModuleCollectRes.Items, data)
}

func (b *BaseCollecter) yasdbVersion() (e error) {
	yasdbVersionItem := data.YtcItem{ItemName: data.BASE_YASDB_VERION}
	defer b.fillResult(&yasdbVersionItem)

	log := log.Module.M(data.BASE_YASDB_VERION)
	yasdbBinPath := path.Join(b.YasdbHome, "bin", bashdef.CMD_YASDB)
	if !fs.IsFileExist(yasdbBinPath) {
		e = &errdef.ErrFileNotFound{Fname: yasdbBinPath}
		log.Error(e)
		yasdbVersionItem.Err = e.Error() // TODO: 补充description信息
		return
	}
	execer := execer.NewExecer(log)
	env := []string{fmt.Sprintf("%s=%s", yasqlutil.LIB_KEY, path.Join(b.YasdbHome, yasqlutil.LIB_PATH))}
	ret, stdout, stderr := execer.EnvExec(env, yasdbBinPath, "-V")
	if ret != 0 {
		e = fmt.Errorf("failed to get yasdb version, err: %s", stderr)
		log.Error(e)
		yasdbVersionItem.Err = stderr
		return
	}
	yasdbVersionItem.Details = strings.TrimSpace(stdout)
	return
}

func (b *BaseCollecter) yasdbPrameter() (e error) {
	yasdbParameterItem := data.YtcItem{ItemName: data.BASE_YASDB_PARAMTER}
	defer b.fillResult(&yasdbParameterItem)

	log := log.Module.M(data.BASE_YASDB_PARAMTER)
	var errs []string // used to collect all errors
	detail := make(map[string]interface{})
	// collect yasdb ini config
	if yasdbIni, e := b.getYasdbIni(); e != nil {
		errs = append(errs, e.Error())
		log.Error(e)
	} else {
		detail[_yasdb_ini_key] = yasdbIni
	}
	if !b.notConnectDB {
		// collect parameter from v$parameter
		if pv, e := b.getYasdbParameter(); e != nil {
			errs = append(errs, e.Error())
			log.Error(e)
		} else {
			detail[_yasdb_parameter_key] = pv
		}
	}
	yasdbParameterItem.Err = strings.Join(errs, stringutil.STR_NEWLINE)
	yasdbParameterItem.Details = detail
	return
}

func (b *BaseCollecter) getYasdbIni() (res map[string]string, e error) {
	iniConfigPath := path.Join(b.YasdbData, "config", "yasdb.ini")
	res = make(map[string]string)
	if !fs.IsFileExist(iniConfigPath) {
		e = &errdef.ErrFileNotFound{Fname: iniConfigPath}
		return
	}
	yasdbConf, e := ini.Load(iniConfigPath)
	if e != nil {
		return
	}
	for _, section := range yasdbConf.Sections() {
		for _, key := range section.Keys() {
			res[key.Name()] = key.String()
		}
	}
	return
}

func (b *BaseCollecter) getYasdbParameter() (pv []*yasdb.VParameter, e error) {
	// collect parameter from v$parameter
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	return yasdb.QueryAllParameter(tx)
}

func (b *BaseCollecter) hostOSInfo() (e error) {
	hostBaseInfoItem := data.YtcItem{ItemName: data.BASE_HOST_OS_INFO}
	defer b.fillResult(&hostBaseInfoItem)

	log := log.Module.M(data.BASE_HOST_OS_INFO)
	hostInfo, e := host.Info()
	if e != nil {
		log.Error(e)
		hostBaseInfoItem.Err = e.Error()
		return
	}
	hostBaseInfoItem.Details = hostInfo
	return
}

func (b *BaseCollecter) hostFirewalldStatus() (e error) {
	hostFirewallStatus := data.YtcItem{ItemName: data.BASE_HOST_FIREWALLD}
	defer b.fillResult(&hostFirewallStatus)

	log := log.Module.M(data.BASE_HOST_FIREWALLD)
	execer := execer.NewExecer(log)
	osRelease, e := osutil.GetOsRelease()
	if e != nil {
		log.Error(e)
		hostFirewallStatus.Err = e.Error()
		return
	}
	if osRelease.Id == osutil.UBUNTU_ID { // ubuntu
		ret, stdout, stderr := execer.Exec(bashdef.CMD_BASH, "-c", bashdef.CMD_UFW, "status")
		if ret != 0 {
			e = fmt.Errorf("failed to get firewalld status, err: %s", stderr)
			hostFirewallStatus.Err = e.Error()
			return
		}
		hostFirewallStatus.Details = strings.Contains(stdout, _ubuntu_firewalld_active)
		return
	}
	// other os
	ret, stdout, stderr := execer.Exec(bashdef.CMD_BASH, "-c", bashdef.CMD_SYSTEMCTL, "is-active", "firewalld")
	if ret != 0 {
		e = fmt.Errorf("failed to get firewalld status, err: %s", stderr)
		hostFirewallStatus.Err = e.Error()
		return
	}
	hostFirewallStatus.Details = strings.Contains(stdout, _firewalld_active) && !strings.Contains(stdout, _firewalld_inactive)
	return
}

func (b *BaseCollecter) hostCPUInfo() (e error) {
	hostCpuInfo := data.YtcItem{ItemName: data.BASE_HOST_CPU}
	defer b.fillResult(&hostCpuInfo)

	log := log.Module.M(data.BASE_HOST_CPU)
	cpuInfo, e := cpu.Info()
	if e != nil {
		log.Error(e)
		hostCpuInfo.Err = e.Error()
		return
	}
	hostCpuInfo.Details = cpuInfo
	return
}

func (b *BaseCollecter) hostDiskInfo() (e error) {
	hostDiskInfo := data.YtcItem{ItemName: data.BASE_HOST_DISK}
	defer b.fillResult(&hostDiskInfo)

	log := log.Module.M(data.BASE_HOST_DISK)
	partitions, e := disk.Partitions(false)
	if e != nil {
		log.Error(e)
		hostDiskInfo.Err = e.Error()
		return
	}
	var usages []disk.UsageStat
	for _, partition := range partitions {
		var usage *disk.UsageStat
		usage, e = disk.Usage(partition.Mountpoint)
		if e != nil {
			log.Error(e)
			hostDiskInfo.Err = e.Error()
			return
		}
		usages = append(usages, *usage)
	}
	hostDiskInfo.Details = usages
	return
}

func (b *BaseCollecter) hostNetworkInfo() (e error) {
	hostNetInfo := data.YtcItem{ItemName: data.BASE_HOST_Network}
	defer b.fillResult(&hostNetInfo)

	log := log.Module.M(data.BASE_HOST_Network)
	netInfo, e := net.Interfaces()
	if e != nil {
		log.Error(e)
		hostNetInfo.Err = e.Error()
		return
	}
	hostNetInfo.Details = netInfo
	return
}

func (b *BaseCollecter) hostMemoryInfo() (e error) {
	hostMemoryInfo := data.YtcItem{ItemName: data.BASE_HOST_Memery}
	defer b.fillResult(&hostMemoryInfo)

	log := log.Module.M(data.BASE_HOST_Memery)
	memInfo, e := mem.VirtualMemory()
	if e != nil {
		log.Error(e)
		hostMemoryInfo.Err = e.Error()
		return
	}
	hostMemoryInfo.Details = memInfo
	return
}

func (b *BaseCollecter) hostCPUUsage() (e error) {
	hostCPUUsage := data.YtcItem{ItemName: data.BASE_HOST_CPU_USAGE}
	defer b.fillResult(&hostCPUUsage)

	log := log.Module.M(data.BASE_HOST_CPU_USAGE)
	data, e := b.hostWorkload(log, data.BASE_HOST_CPU_USAGE)
	if e != nil {
		hostCPUUsage.Err = e.Error()
		return
	}
	hostCPUUsage.Details = data
	return
}

func (b *BaseCollecter) hostNetworkIO() (e error) {
	hostNetworkIO := data.YtcItem{ItemName: data.BASE_HOST_NETWORK_IO}
	defer b.fillResult(&hostNetworkIO)

	log := log.Module.M(data.BASE_HOST_NETWORK_IO)
	data, e := b.hostWorkload(log, data.BASE_HOST_NETWORK_IO)
	if e != nil {
		hostNetworkIO.Err = e.Error()
		return
	}
	hostNetworkIO.Details = data
	return
}

func (b *BaseCollecter) hostDiskIO() (e error) {
	hostDiskIO := data.YtcItem{ItemName: data.BASE_HOST_DISK_IO}
	defer b.fillResult(&hostDiskIO)

	log := log.Module.M(data.BASE_HOST_DISK_IO)
	data, e := b.hostWorkload(log, data.BASE_HOST_DISK_IO)
	if e != nil {
		hostDiskIO.Err = e.Error()
		return
	}
	hostDiskIO.Details = data
	return
}

func (b *BaseCollecter) hostMemoryUsage() (e error) {
	hostMemoryUsage := data.YtcItem{ItemName: data.BASE_HOST_MEMORY_USAGE}
	defer b.fillResult(&hostMemoryUsage)

	log := log.Module.M(data.BASE_HOST_MEMORY_USAGE)
	data, e := b.hostWorkload(log, data.BASE_HOST_MEMORY_USAGE)
	if e != nil {
		hostMemoryUsage.Err = e.Error()
		return
	}
	hostMemoryUsage.Details = data
	return
}

func (b *BaseCollecter) hostWorkload(log yaslog.YasLog, itemName string) (data interface{}, e error) {
	details := map[string]interface{}{}
	hasSar := b.CheckSarAccess() == nil
	if hasSar { // collect historyworkload
		historyNetworkWorkload, err := b.hostHistoryWorkload(log, itemName, b.StartTime, b.EndTime)
		if err != nil {
			e = fmt.Errorf("failed to collect history %s, err: %s", itemName, err.Error())
			log.Error(err)
		} else {
			details[_history_key] = historyNetworkWorkload
		}
	}
	// collect current workload
	currentNetworkWorkload, err := b.hostCurrentWorkload(log, itemName, hasSar)
	if err != nil {
		e = fmt.Errorf("failed to collect current %s, err: %s", itemName, err.Error())
		log.Error(err)
	} else {
		details[_current_key] = currentNetworkWorkload
	}
	data = details
	return
}

func (b *BaseCollecter) hostHistoryWorkload(log yaslog.YasLog, itemName string, start, end time.Time) (interface{}, error) {
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err := fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return nil, err
	}
	sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
	if !ok {
		err := fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
		log.Error(err)
		return nil, err
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
		output, err := sar.Collect(workloadType, sarArg, arg)
		if err != nil {
			log.Error(err)
			continue
		}
		for timestamp, output := range output {
			sarOutput[timestamp] = output
		}
	}
	return sarOutput, nil
}

// TODO:这里的时区要需要处理
func (b *BaseCollecter) genHistoryWorkloadArgs(start, end time.Time, sarDir string) []string {
	// get data between start and end
	var dates []time.Time
	begin := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	for date := begin; !date.After(end); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date)
	}
	args := []string{} // used to collect history data
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
	return args
}

func (b *BaseCollecter) hostCurrentWorkload(log yaslog.YasLog, itemName string, hasSar bool) (interface{}, error) {
	// global conf
	strategyConf := confdef.GetStrategyConf()
	scrapeInterval, scrapeTimes := strategyConf.Collect.ScrapeInterval, strategyConf.Collect.ScrapeTimes
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err := fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return nil, err
	}
	if hasSar { // use sar to collect first
		sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
		if !ok {
			err := fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
			log.Error(err)
			return nil, err
		}
		sar := sar.NewSar(log)
		return sar.Collect(workloadType, sarArg, strconv.Itoa(scrapeInterval), strconv.Itoa(scrapeTimes))
	}
	// use gopsutil to calculate by ourself
	return gopsutil.Collect(workloadType, scrapeInterval, scrapeTimes)
}
