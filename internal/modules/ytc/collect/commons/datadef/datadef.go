package datadef

const (
	// base info
	BASE_YASDB_VERION      = "YashanDB-Version"
	BASE_YASDB_PARAMETER   = "YashanDB-Parameter"
	BASE_HOST_OS_INFO      = "Host-OSInfo"
	BASE_HOST_FIREWALLD    = "Host-FirewalldStatus"
	BASE_HOST_CPU          = "Host-CPU"
	BASE_HOST_DISK         = "Host-Disk"
	BASE_HOST_NETWORK      = "Host-Network"
	BASE_HOST_MEMORY       = "Host-Memory"
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
	PERF_YASDB_AWR      = "YashanDB-AWR"
	PERF_YASDB_SLOW_SQL = "YashanDB-SlowSQL"

	// extra file collect
	EXTRA_FILE_COLLECT = "Extra-FileCollect"
)
