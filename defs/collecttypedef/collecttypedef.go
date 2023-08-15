package collecttypedef

import (
	"errors"
	"time"
)

const (
	TYPE_BASE = "base"
	TYPE_DIAG = "diag"
	TYPE_PREF = "pref"
)

const (
	WT_CPU     WorkloadType = "cpu"
	WT_NETWORK WorkloadType = "network"
	WT_MEMORY  WorkloadType = "memory"
	WT_DISK    WorkloadType = "disk"
)

var (
	ErrKnownType = errors.New("unknow collect type")
)

type CollectParam struct {
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	Output        string    `json:"output"`
	YasdbHome     string    `json:"yasdbHome"`
	YasdbData     string    `json:"yasdbData"`
	YasdbUser     string    `json:"yasdbUser"`
	YasdbPassword string    `json:"yasdbPassword"`
}

type WorkloadItem map[string]interface{}

type WorkloadOutput map[int64]WorkloadItem

type WorkloadType string
