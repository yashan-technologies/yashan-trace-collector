package report

import (
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/data/reporter/baseinforeporter"
	"ytc/internal/modules/ytc/collect/data/reporter/commons"
	"ytc/internal/modules/ytc/collect/data/reporter/diagreporter"
	"ytc/internal/modules/ytc/collect/data/reporter/extrareporter"
	"ytc/internal/modules/ytc/collect/data/reporter/performancereporter"
)

var REPORTERS = map[string]commons.Reporter{
	// BASE
	datadef.BASE_YASDB_VERION:      baseinforeporter.NewYashanDBVersionReporter(),
	datadef.BASE_YASDB_PARAMETER:   baseinforeporter.NewYashanDBParameterReporter(),
	datadef.BASE_HOST_OS_INFO:      baseinforeporter.NewHostOSInfoReporter(),
	datadef.BASE_HOST_FIREWALLD:    baseinforeporter.NewHostFirewallReporterReporter(),
	datadef.BASE_HOST_CPU:          baseinforeporter.NewHostCPUReporter(),
	datadef.BASE_HOST_DISK:         baseinforeporter.NewHostDiskReporter(),
	datadef.BASE_HOST_NETWORK:      baseinforeporter.NewHostNetworkReporter(),
	datadef.BASE_HOST_MEMORY:       baseinforeporter.NewHostMemoryReporter(),
	datadef.BASE_HOST_NETWORK_IO:   baseinforeporter.NewHostNetworkIOReporter(),
	datadef.BASE_HOST_CPU_USAGE:    baseinforeporter.NewHostCPUUsageReporter(),
	datadef.BASE_HOST_DISK_IO:      baseinforeporter.NewHostDiskIOReporter(),
	datadef.BASE_HOST_MEMORY_USAGE: baseinforeporter.NewHostMemoryUsageReporter(),

	// DIAG
	datadef.DIAG_YASDB_PROCESS_STATUS:  diagreporter.NewYashanDBProcessStatusReporter(),
	datadef.DIAG_YASDB_INSTANCE_STATUS: diagreporter.NewYashanDBInstanceStatusReporter(),
	datadef.DIAG_YASDB_DATABASE_STATUS: diagreporter.NewYashanDBDatabaseStatusReporter(),
	datadef.DIAG_YASDB_ADR:             diagreporter.NewYashanDBADRLogReporter(),
	datadef.DIAG_YASDB_RUNLOG:          diagreporter.NewYashanDBRunLogReporter(),
	datadef.DIAG_YASDB_ALERTLOG:        diagreporter.NewYashanDBAlertLogReporter(),
	datadef.DIAG_YASDB_COREDUMP:        diagreporter.NewYashanDBCoreDumpReporter(),
	datadef.DIAG_HOST_SYSTEMLOG:        diagreporter.NewHostSystemLogReporter(),
	datadef.DIAG_HOST_KERNELLOG:        diagreporter.NewHostKernelLogReporter(),
	datadef.DIAG_HOST_BASH_HISTORY:     diagreporter.NewHostBashHistoryReporter(),

	// PERF
	datadef.PERF_YASDB_AWR:      performancereporter.NewAWRReporter(),
	datadef.PERF_YASDB_SLOW_SQL: performancereporter.NewSlowSqlReporter(),

	// EXTR
	datadef.EXTRA_FILE_COLLECT: extrareporter.NewExtraFileReporter(),
}
