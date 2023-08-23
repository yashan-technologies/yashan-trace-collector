package baseinfo

import (
	"fmt"
	"path"
	"strconv"
	"time"

	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	"ytc/defs/confdef"
	"ytc/defs/timedef"
	"ytc/internal/modules/ytc/collect/baseinfo/gopsutil"
	"ytc/internal/modules/ytc/collect/baseinfo/sar"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/utils/stringutil"

	"git.yasdb.com/go/yaslog"
)

type HostWorkResponse struct {
	Data     map[string]interface{}
	Errors   map[string]string
	DataType datadef.DataType
}

func (b *BaseCollecter) hostWorkload(log yaslog.YasLog, itemName string) (resp HostWorkResponse, err error) {
	details := map[string]interface{}{}
	hasSar := b.CheckSarAccess() == nil
	resp.DataType = datadef.DATATYPE_GOPSUTIL
	resp.Errors = make(map[string]string)

	// collect historyworkload
	if hasSar {
		resp.DataType = datadef.DATATYPE_SAR
		if historyNetworkWorkload, e := b.hostHistoryWorkload(log, itemName, b.StartTime, b.EndTime); e != nil {
			err = fmt.Errorf("failed to collect history %s, err: %s", itemName, e.Error())
			resp.Errors[KEY_HISTORY] = err.Error()
			log.Error(err)
		} else {
			details[KEY_HISTORY] = historyNetworkWorkload
		}
	} else {
		err = fmt.Errorf("cannot find command '%s'", bashdef.CMD_SAR)
		resp.Errors[KEY_HISTORY] = err.Error()
		log.Error(err)
	}

	// collect current workload
	if currentNetworkWorkload, e := b.hostCurrentWorkload(log, itemName, hasSar); e != nil {
		err = fmt.Errorf("failed to collect current %s, err: %s", itemName, e.Error())
		resp.Errors[KEY_CURRENT] = err.Error()
		log.Error(err)
	} else {
		details[KEY_CURRENT] = currentNetworkWorkload
	}
	resp.Data = details
	return
}

func (b *BaseCollecter) hostHistoryWorkload(log yaslog.YasLog, itemName string, start, end time.Time) (resp collecttypedef.WorkloadOutput, err error) {
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err = fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return
	}
	sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
	if !ok {
		err = fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
		log.Error(err)
		return
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
		output, e := sar.Collect(workloadType, sarArg, arg)
		if e != nil {
			log.Error(e)
			continue
		}
		for timestamp, output := range output {
			sarOutput[timestamp] = output
		}
	}
	resp = sarOutput
	return
}

func (b *BaseCollecter) genHistoryWorkloadArgs(start, end time.Time, sarDir string) (args []string) {
	// get data between start and end
	var dates []time.Time
	begin := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	for date := begin; !date.After(end); date = date.AddDate(0, 0, 1) {
		dates = append(dates, date)
	}
	for i, date := range dates {
		var startArg, endArg, fileArg string
		// the frist
		if i == 0 && !date.Equal(start) {
			startArg = fmt.Sprintf("-s %s", start.Format(timedef.TIME_FORMAT_TIME))
		}
		// the last one
		if i == len(dates)-1 {
			if date.Equal(end) {
				// skip
				continue
			}
			endArg = fmt.Sprintf("-e %s", end.Format(timedef.TIME_FORMAT_TIME))
		}
		fileArg = fmt.Sprintf("-f %s", path.Join(sarDir, fmt.Sprintf("sa%s", date.Format(timedef.TIME_FORMAT_DAY))))
		args = append(args, fmt.Sprintf("%s %s %s", fileArg, startArg, endArg))
	}
	return
}

func (b *BaseCollecter) hostCurrentWorkload(log yaslog.YasLog, itemName string, hasSar bool) (resp collecttypedef.WorkloadOutput, err error) {
	// global conf
	strategyConf := confdef.GetStrategyConf()
	scrapeInterval, scrapeTimes := strategyConf.Collect.ScrapeInterval, strategyConf.Collect.ScrapeTimes
	// get sar args
	workloadType, ok := ItemNameToWorkloadTypeMap[itemName]
	if !ok {
		err = fmt.Errorf("failed to get workload type from item name: %s", itemName)
		log.Error(err)
		return
	}
	// use sar to collect first
	if hasSar {
		sarArg, ok := WorkloadTypeToSarArgMap[workloadType]
		if !ok {
			err = fmt.Errorf("failed to get SAR arg from workload type: %s", workloadType)
			log.Error(err)
			return
		}
		sar := sar.NewSar(log)
		return sar.Collect(workloadType, sarArg, strconv.Itoa(scrapeInterval), strconv.Itoa(scrapeTimes))
	}
	// use gopsutil to calculate by ourself
	return gopsutil.Collect(workloadType, scrapeInterval, scrapeTimes)
}
