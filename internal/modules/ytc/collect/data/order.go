package data

import (
	"ytc/defs/collecttypedef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
)

// module ordera
var (
	_moduleOrder = []string{
		collecttypedef.TYPE_BASE,
		collecttypedef.TYPE_DIAG,
		collecttypedef.TYPE_EXTRA,
		collecttypedef.TYPE_PERF,
	}
)

// item ordera
var (
	_baseItemOrder = []string{
		datadef.BASE_YASDB_VERION,
		datadef.BASE_YASDB_PARAMETER,
		datadef.BASE_HOST_OS_INFO,
		datadef.BASE_HOST_FIREWALLD,
		datadef.BASE_HOST_CPU,
		datadef.BASE_HOST_MEMORY,
		datadef.BASE_HOST_DISK,
		datadef.BASE_HOST_NETWORK,
		datadef.BASE_HOST_CPU_USAGE,
		datadef.BASE_HOST_MEMORY_USAGE,
		datadef.BASE_HOST_NETWORK_IO,
		datadef.BASE_HOST_DISK_IO,
	}

	_diagItemOrder = []string{
		datadef.DIAG_YASDB_PROCESS_STATUS,
		datadef.DIAG_YASDB_INSTANCE_STATUS,
		datadef.DIAG_YASDB_DATABASE_STATUS,
		datadef.DIAG_YASDB_COREDUMP,
		datadef.DIAG_YASDB_RUNLOG,
		datadef.DIAG_YASDB_ALERTLOG,
		datadef.DIAG_YASDB_ADR,
		datadef.DIAG_HOST_SYSTEMLOG,
		datadef.DIAG_HOST_KERNELLOG,
	}

	// TODO: add items to perf item order
	_perfItemOrder = []string{
		datadef.PERF_YASDB_AWR,
		datadef.PERF_YASDB_SLOW_SQL,
	}

	_extraItemOrder = []string{
		datadef.EXTRA_FILE_COLLECT,
	}

	_itemOrder = map[string][]string{
		collecttypedef.TYPE_BASE:  _baseItemOrder,
		collecttypedef.TYPE_DIAG:  _diagItemOrder,
		collecttypedef.TYPE_EXTRA: _extraItemOrder,
		collecttypedef.TYPE_PERF:  _perfItemOrder,
	}
)
