package data

import (
	"strings"
	"sync"
	"time"

	"ytc/defs/collecttypedef"

	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/resultgenner"
	"ytc/utils/stringutil"
)

const (
	// base info
	BASE_YASDB_VERION      = "YashanDB-Version"
	BASE_YASDB_PARAMTER    = "YashanDB-Paramter"
	BASE_HOST_OS_INFO      = "Host-OSInfo"
	BASE_HOST_FIREWALLD    = "Host-FirewalldStatus"
	BASE_HOST_CPU          = "Host-CPU"
	BASE_HOST_DISK         = "Host-Disk"
	BASE_HOST_Network      = "Host-Network"
	BASE_HOST_Memery       = "Host-Memory"
	BASE_HOST_NETWORK_IO   = "Host-NetworkIO"
	BASE_HOST_CPU_USAGE    = "Host-CPUUsage"
	BASE_HOST_DISK_IO      = "Host-DiskIO"
	BASE_HOST_MEMORY_USAGE = "Host-MemoryUsage"

	// diagnosis info
	DIAG_YASDB_PROCESS_STATUS  = "YashanDB-ProcessStatus"
	DIAG_YASDB_INSTANCE_STATUS = "YashanDB-InstanceStatus"
	DIAG_YASDB_DATABASE_STATUS = "YashanDB-DatabaseStatus"
	DIAG_YASDB_ADR             = "YashanDB-ADR"
	DIAG_YASDB_RUNLOG          = "YashanDB-RunLog"
	DIAG_YASDB_ALERTLOG        = "YashanDB-AlertLog"
	DIAG_YASDB_COREDUMP        = "YashanDB-Coredump"
	DIAG_HOST_KERNELLOG        = "Host-KernelLog"
	DIAG_HOST_SYSTEMLOG        = "Host-SystemLog"
	DIAG_HOST_DMESG            = "Host-Dmesg"

	// performance info

	// extra file collect
	EXTRA_FILE_COLLECT = "Extra-FileCollect"
)

const (
	TXT_REPORT  = "txt"
	MD_REPORT   = "md"
	HTML_REPORY = "html"
)

type NoAccessRes struct {
	ModuleItem   string
	Description  string
	Tips         string
	ForceCollect bool // default false
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type YtcItem struct {
	ItemName    string      `json:"itemName"`    // item name
	Err         string      `json:"err"`         // 原始报错信息
	Description string      `json:"description"` // 失败原因描述
	Details     interface{} `json:"details"`     // 每个收集项包含的数据
}

type YtcModule struct {
	sync.Mutex
	Module string     `json:"module"`
	Items  []*YtcItem `json:"items"`
}

type YtcReport struct {
	sync.Mutex
	CollectBeginTime time.Time                    `json:"collectBeginTime"`
	CollectEndtime   time.Time                    `json:"collectEndTime"`
	CollectParam     *collecttypedef.CollectParam `json:"collectParam"`
	ModuleResults    map[string]*YtcModule        `json:"moduleResults"`
	genner           resultgenner.BaseGenner
}

func NewYtcReport(param *collecttypedef.CollectParam) *YtcReport {
	return &YtcReport{
		CollectParam:  param,
		ModuleResults: make(map[string]*YtcModule),
		genner:        resultgenner.BaseGenner{},
	}
}

// [Interface Func]
func (c *YtcReport) GenData(data interface{}, fname string) error {
	return c.genner.GenData(data, fname)
}

// [Interface Func]
func (c *YtcReport) GenReport() []byte {
	var content string
	content = strings.TrimSuffix(content, stringutil.STR_NEWLINE)
	return []byte(content)
}

func (c *YtcReport) GenResult(outputDir, reportType string, types map[string]struct{}) (string, error) {
	genner := resultgenner.BaseResultGenner{
		Datas:        c.ModuleResults,
		CollectTypes: types,
		OutputDir:    outputDir,
		ReportType:   reportType,
		Timestamp:    c.CollectBeginTime.Format(timedef.TIME_FORMAT_IN_FILE),
		Genner:       c,
	}
	return genner.GenResult()
}

func (c *YtcReport) GetPackageDir() string {
	genner := resultgenner.BaseResultGenner{
		OutputDir: c.CollectParam.Output,
		Timestamp: c.CollectBeginTime.Format(timedef.TIME_FORMAT_IN_FILE),
	}
	return genner.GetPackageDir()
}
