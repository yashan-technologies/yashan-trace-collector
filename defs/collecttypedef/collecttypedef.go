package collecttypedef

import (
	"errors"
	"fmt"
	"time"

	"ytc/defs/timedef"
)

const (
	TYPE_BASE  = "base"
	TYPE_DIAG  = "diag"
	TYPE_PERF  = "perf"
	TYPE_EXTRA = "extra"
)

const (
	WT_CPU     WorkloadType = "cpu"
	WT_NETWORK WorkloadType = "network"
	WT_MEMORY  WorkloadType = "memory"
	WT_DISK    WorkloadType = "disk"
)

const PACKAGE_NAME_PREFIX = "ytc"

var (
	ErrKnownType = errors.New("unknow collect type")
)

var (
	typeFullName = map[string]string{
		TYPE_BASE: "baseinfo",
		TYPE_DIAG: "diagnosis",
		TYPE_PERF: "performance",
	}
)

var (
	CollectTypeChineseName = map[string]string{
		TYPE_BASE:  "基础信息",
		TYPE_DIAG:  "诊断信息",
		TYPE_PERF:  "性能调优信息",
		TYPE_EXTRA: "额外收集项",
	}
)

type CollectParam struct {
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	Output          string    `json:"output"`
	YasdbHome       string    `json:"yasdbHome"`
	YasdbData       string    `json:"yasdbData"`
	YasdbUser       string    `json:"yasdbUser"`
	YasdbPassword   string    `json:"yasdbPassword"`
	Include         []string  `json:"include"`
	Exclude         []string  `json:"exclude"`
	BeginTime       time.Time `json:"-"`
	YasdbHomeOSUser string    `json:"-"`
}

type WorkloadItem map[string]interface{}

type WorkloadOutput map[int64]WorkloadItem

type WorkloadType string

func GetTypeFullName(s string) string {
	full, ok := typeFullName[s]
	if !ok {
		full = s
	}
	return full
}

func (c *CollectParam) GetPackageTimestamp() string {
	// use begin time as timestamp
	return c.BeginTime.Format(timedef.TIME_FORMAT_IN_FILE)
}

func (c *CollectParam) GetPackageName() string {
	return fmt.Sprintf("%s-%s", PACKAGE_NAME_PREFIX, c.GetPackageTimestamp())
}
