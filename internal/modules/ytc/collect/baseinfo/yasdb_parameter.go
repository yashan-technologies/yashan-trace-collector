package baseinfo

import (
	"path"

	"ytc/defs/errdef"
	"ytc/internal/modules/ytc/collect/commons/datadef"
	"ytc/internal/modules/ytc/collect/yasdb"
	"ytc/log"
	"ytc/utils/yasqlutil"

	"git.yasdb.com/go/yasutil/fs"
	ini "gopkg.in/ini.v1"
)

func (b *BaseCollecter) getYasdbParameter() (err error) {
	yasdbParameterItem := datadef.YTCItem{
		Name:     datadef.BASE_YASDB_PARAMETER,
		Children: make(map[string]datadef.YTCItem),
	}
	defer b.fillResult(&yasdbParameterItem)
	log := log.Module.M(datadef.BASE_YASDB_PARAMETER)

	// collect yasdb ini config
	if yasdbIni, err := b.getYasdbIni(); err != nil {
		yasdbParameterItem.Children[KEY_YASDB_INI] = datadef.YTCItem{Error: err.Error(), Description: datadef.GenDefaultDesc()}
		log.Errorf("failed to get yasdb.ini, err: %s", err.Error())
	} else {
		yasdbParameterItem.Children[KEY_YASDB_INI] = datadef.YTCItem{Details: yasdbIni}
	}

	// collect parameter from v$parameter
	if !b.notConnectDB {
		if pv, err := b.getParameter(); err != nil {
			yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Error: err.Error(), Description: datadef.GenDefaultDesc()}
			log.Errorf("failed to get yashandb parameter, err: %s", err.Error())
		} else {
			yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Details: pv}
		}
	} else {
		yasdbParameterItem.Children[KEY_YASDB_PARAMETER] = datadef.YTCItem{Error: "cannot connect to database", Description: datadef.GenSkipCollectDatabaseInfoDesc()}
	}
	return
}

func (b *BaseCollecter) getYasdbIni() (res map[string]string, err error) {
	iniConfigPath := path.Join(b.YasdbData, CONFIG_DIR_NAME, KEY_YASDB_INI)
	res = make(map[string]string)
	if !fs.IsFileExist(iniConfigPath) {
		err = &errdef.ErrFileNotFound{Fname: iniConfigPath}
		return
	}
	yasdbConf, err := ini.Load(iniConfigPath)
	if err != nil {
		return
	}
	for _, section := range yasdbConf.Sections() {
		for _, key := range section.Keys() {
			res[key.Name()] = key.String()
		}
	}
	return
}

func (b *BaseCollecter) getParameter() (pv []*yasdb.VParameter, err error) {
	// collect parameter from v$parameter
	tx := yasqlutil.GetLocalInstance(b.YasdbUser, b.YasdbPassword, b.YasdbHome, b.YasdbData)
	return yasdb.QueryAllParameter(tx)
}
