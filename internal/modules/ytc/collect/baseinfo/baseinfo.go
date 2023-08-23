package baseinfo

import (
	"ytc/defs/collecttypedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"
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

const (
	CONFIG_DIR_NAME = "config"
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

func (b *BaseCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.BASE_YASDB_VERION:      b.getYasdbVersion,
		datadef.BASE_YASDB_PARAMETER:   b.getYasdbParameter,
		datadef.BASE_HOST_OS_INFO:      b.getHostOSInfo,
		datadef.BASE_HOST_FIREWALLD:    b.getHostFirewalldStatus,
		datadef.BASE_HOST_CPU:          b.getHostCPUInfo,
		datadef.BASE_HOST_DISK:         b.getHostDiskInfo,
		datadef.BASE_HOST_NETWORK:      b.getHostNetworkInfo,
		datadef.BASE_HOST_MEMORY:       b.getHostMemoryInfo,
		datadef.BASE_HOST_NETWORK_IO:   b.getHostNetworkIO,
		datadef.BASE_HOST_CPU_USAGE:    b.getHostCPUUsage,
		datadef.BASE_HOST_DISK_IO:      b.getHostDiskIO,
		datadef.BASE_HOST_MEMORY_USAGE: b.getHostMemoryUsage,
	}
}

// [Interface Func]
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
func (b *BaseCollecter) ItemsToCollect(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
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

// [Interface Func]
func (b *BaseCollecter) PreCollect(packageDir string) (err error) {
	return
}

// [Interface Func]
func (b *BaseCollecter) CollectOK() *datadef.YTCModule {
	return b.ModuleCollectRes
}

func (b *BaseCollecter) fillResult(data *datadef.YTCItem) {
	b.ModuleCollectRes.Set(data)
}
