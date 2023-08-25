package diagnosis

import (
	"path"
	"time"

	"ytc/defs/collecttypedef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/log"

	"git.yasdb.com/go/yasutil/fs"
)

type checkFunc func() *ytccollectcommons.NoAccessRes

const (
	CORE_PATTERN_PATH = "/proc/sys/kernel/core_pattern"

	ABRT_HOOK_CPP         = "abrt-hook-ccpp"
	ABRT_CONF             = "/etc/abrt/abrt.conf"
	KEY_DUMP_LOCATION     = "DumpLocation"
	DEFAULT_DUMP_LOCATION = "/var/spool/abrt"

	SYSTEMD_COREDUMP         = "systemd-coredump"
	SYSTEMD_COREDUMP_CONF    = "/etc/systemd/coredump.conf"
	KEY_STORAGE              = "Storage"
	VALUE_EXTERNAL           = "external"
	DEFAULT_EXTERNAL_STORAGE = "/var/lib/systemd/coredump"
	LOG_ROTATE_CONFIG        = "/etc/logrotate.conf"

	SYSTEM_LOG_MESSAGES = "/var/log/messages"
	SYSTEM_LOG_SYSLOG   = "/var/log/syslog"

	DIAG_DIR_NAME  = "diag"
	LOG_DIR_NAME   = "log"
	EXTRA_DIR_NAME = "extra"

	CORE_DUMP_DIR_NAME = "coredump"
	ADR_DIR_NAME       = "adr"

	YASDB_DIR_NAME  = "yasdb"
	SYSTEM_DIR_NAME = "system"

	YASDB_ALERT_LOG = "alert"
	YASDB_RUN_LOG   = "run"

	SYSTEM_DMESG_LOG    = "dmesg"
	SYSTEM_MESSAGES_LOG = "messages"
	SYSTEM_SYS_LOG      = "syslog"

	LOG_FILE_SUFFIX = "%s.log"
	TAR_FILE_SUFFIX = "%s.tar.gz"

	CORE_FILE_KEY = "core"
)

var (
	DiagChineseName = map[string]string{
		datadef.DIAG_YASDB_ADR:             "数据库ADR日志",
		datadef.DIAG_YASDB_RUNLOG:          "数据库run.log日志",
		datadef.DIAG_YASDB_ALERTLOG:        "数据库alert.log日志",
		datadef.DIAG_YASDB_PROCESS_STATUS:  "数据库进程信息",
		datadef.DIAG_YASDB_INSTANCE_STATUS: "数据库实例状态",
		datadef.DIAG_YASDB_DATABASE_STATUS: "数据库状态",
		datadef.DIAG_HOST_SYSTEMLOG:        "操作系统日志",
		datadef.DIAG_HOST_KERNELLOG:        "操作系统内核日志",
		datadef.DIAG_YASDB_COREDUMP:        "CoreDump",
	}
)

var _packageDir = ""

type logTimeParseFunc func(date time.Time, line string) (time.Time, error)

type DiagCollecter struct {
	*collecttypedef.CollectParam
	ModuleCollectRes *datadef.YTCModule
	yasdbValidateErr error
	notConnectDB     bool
}

func NewDiagCollecter(collectParam *collecttypedef.CollectParam) *DiagCollecter {
	return &DiagCollecter{
		CollectParam: collectParam,
		ModuleCollectRes: &datadef.YTCModule{
			Module: collecttypedef.TYPE_DIAG,
		},
	}
}

func (b *DiagCollecter) itemFunc() map[string]func() error {
	return map[string]func() error{
		datadef.DIAG_YASDB_PROCESS_STATUS:  b.getYasdbProcessStatus,
		datadef.DIAG_YASDB_INSTANCE_STATUS: b.getYasdbInstanceStatus,
		datadef.DIAG_YASDB_DATABASE_STATUS: b.getYasdbDatabaseStatus,
		datadef.DIAG_YASDB_ADR:             b.collectYasdbADR,
		datadef.DIAG_YASDB_ALERTLOG:        b.collectYasdbAlertLog,
		datadef.DIAG_YASDB_RUNLOG:          b.collectYasdbRunLog,
		datadef.DIAG_YASDB_COREDUMP:        b.yasdbCoreDumpFile,
		datadef.DIAG_HOST_SYSTEMLOG:        b.collectHostSystemLog,
		datadef.DIAG_HOST_KERNELLOG:        b.collectHostKernelLog,
	}
}

// [Interface Func]
func (d *DiagCollecter) CheckAccess(yasdbValidate error) (noAccess []ytccollectcommons.NoAccessRes) {
	d.yasdbValidateErr = yasdbValidate
	noAccess = make([]ytccollectcommons.NoAccessRes, 0)
	funcMap := d.CheckFunc()
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
func (d *DiagCollecter) CollectFunc(items []string) (res map[string]func() error) {
	res = make(map[string]func() error)
	itemFuncMap := d.itemFunc()
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
func (b *DiagCollecter) Type() string {
	return collecttypedef.TYPE_DIAG
}

// [Interface Func]
func (b *DiagCollecter) ItemsToCollect(noAccess []ytccollectcommons.NoAccessRes) (res []string) {
	noMap := b.getNotAccessItem(noAccess)
	for item := range DiagChineseName {
		if _, ok := noMap[item]; !ok {
			res = append(res, item)
		}
	}
	return
}

func (b *DiagCollecter) getNotAccessItem(noAccess []ytccollectcommons.NoAccessRes) (res map[string]struct{}) {
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
func (b *DiagCollecter) PreCollect(packageDir string) (err error) {
	b.setPackageDir(packageDir)
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, CORE_DUMP_DIR_NAME)); err != nil {
		return
	}
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, ADR_DIR_NAME)); err != nil {
		return
	}
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, YASDB_DIR_NAME)); err != nil {
		return
	}
	if err = fs.Mkdir(path.Join(_packageDir, DIAG_DIR_NAME, LOG_DIR_NAME, SYSTEM_DIR_NAME)); err != nil {
		return
	}
	return
}

func (b *DiagCollecter) setPackageDir(packageDir string) {
	_packageDir = packageDir
	log.Module.Infof("package dir is %s", _packageDir)
}

// [Interface Func]
func (b *DiagCollecter) CollectOK() *datadef.YTCModule {
	return b.ModuleCollectRes
}

func (b *DiagCollecter) fillResult(data *datadef.YTCItem) {
	b.ModuleCollectRes.Set(data)
}
