package performance

import (
	"fmt"
	"path"
	"strings"

	"ytc/defs/confdef"
	ytccollectcommons "ytc/internal/modules/ytc/collect/commons"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/fileutil"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yaslog"
)

const (
	USER_SYS = "SYS"
)

func (p *PerfCollecter) checkDatabaseOpenMode(logger yaslog.YasLog) (bool, error) {
	tx := yasqlutil.GetLocalInstance(p.YasdbUser, p.YasdbPassword, p.YasdbHome, p.YasdbData)
	database, err := yasdb.QueryDatabase(tx)
	if err != nil {
		logger.Errorf("query v$database failed: %s", err)
		return false, err
	}
	return database.IsDatabaseInReadWwiteMode(), nil
}

func (p *PerfCollecter) checkAWR() *ytccollectcommons.NoAccessRes {
	noAccess := &ytccollectcommons.NoAccessRes{ModuleItem: datadef.PERF_YASDB_AWR}
	if strings.ToUpper(p.YasdbUser) != USER_SYS {
		ytccollectcommons.FillDescTips(noAccess, fmt.Sprintf(ytccollectcommons.USER_NOT_SYS_DESC, p.YasdbUser), ytccollectcommons.USER_NOT_SYS_TIPS)
		return noAccess
	}
	if p.yasdbValidateErr != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(p.yasdbValidateErr)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}

	logger := log.Module.M(datadef.PERF_YASDB_AWR)
	inReadWwiteMode, err := p.checkDatabaseOpenMode(logger)
	if err != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	if !inReadWwiteMode {
		desc, tips := ytccollectcommons.AWR_SKIP_TIPS, ytccollectcommons.AWR_SKIP_TIPS
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}

	if _, _, err := p.genStartEndSnapId(logger); err != nil {
		if err == yasdb.ErrNoSatisfiedSnapshot {
			desc := ytccollectcommons.NO_SATISFIED_SNAP_DESC
			tips := ytccollectcommons.NO_SATISFIED_TIPS
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		desc, tips := ytccollectcommons.YasErrDescAndtips(err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	collectConfig := confdef.GetStrategyConf().Collect
	timeout := collectConfig.GetAWRTimeout()
	desc := ytccollectcommons.AWR_TIMEOUT_DESC
	tips := fmt.Sprintf(ytccollectcommons.AWR_TIMEOUT_TIPS, timeout.String())
	ytccollectcommons.FillDescTips(noAccess, desc, tips)
	noAccess.ForceCollect = true
	return noAccess
}

func (p *PerfCollecter) checkSlowSql() *ytccollectcommons.NoAccessRes {
	noAccess := &ytccollectcommons.NoAccessRes{ModuleItem: datadef.PERF_YASDB_SLOW_SQL}
	defaultSlowLog := path.Join(p.YasdbData, ytccollectcommons.LOG, ytccollectcommons.SLOW, ytccollectcommons.SLOW_LOG)
	defaultSlowLogTips := fmt.Sprintf(ytccollectcommons.DEFAULT_SLOWSQL_TIPS, defaultSlowLog)
	if p.yasdbValidateErr != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(p.yasdbValidateErr)
		if err := fileutil.CheckAccess(defaultSlowLog); err != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccollectcommons.FillDescTips(noAccess, desc, defaultSlowLogTips)
		noAccess.ForceCollect = true
		return noAccess
	}
	slowLogPath, err := p.getFileSlowLogPath()
	slowLog := path.Join(slowLogPath, ytccollectcommons.SLOW_LOG)
	if err != nil {
		desc, tips := ytccollectcommons.YasErrDescAndtips(err)
		if err := fileutil.CheckAccess(defaultSlowLog); err != nil {
			ytccollectcommons.FillDescTips(noAccess, desc, tips)
			return noAccess
		}
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		noAccess.ForceCollect = true
		return noAccess
	}
	if err := fileutil.CheckAccess(slowLog); err != nil {
		desc, tips := ytccollectcommons.PathErrDescAndTips(slowLog, err)
		ytccollectcommons.FillDescTips(noAccess, desc, tips)
		return noAccess
	}
	return nil
}
