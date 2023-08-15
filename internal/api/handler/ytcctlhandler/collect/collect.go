package ytcctlhandler

import (
	"fmt"
	"strings"
	"time"
	"ytc/defs/bashdef"
	"ytc/defs/collecttypedef"
	ytccollect "ytc/internal/modules/ytc/collect"
	"ytc/internal/modules/ytc/collect/data"
	"ytc/log"
	"ytc/utils/terminalutil/barutil"

	"git.yasdb.com/go/yasutil/tabler"
)

func (c *CollecterHandler) Collect() error {
	noAccessMap, err := c.checkAccess()
	if err != nil {
		log.Handler.Errorf(err.Error())
		return err
	}
	moduleItems, err := c.getCollectItem(noAccessMap)
	if err != nil {
		log.Handler.Errorf(err.Error())
		return err
	}
	if err := c.printCollectItem(moduleItems); err != nil {
		log.Handler.Errorf(err.Error())
		return err
	}
	fmt.Printf("\nStarting collect...\n")
	return c.collect(moduleItems)
}

func (c *CollecterHandler) checkAccess() (map[string][]data.NoAccessRes, error) {
	m := make(map[string][]data.NoAccessRes)
	for _, c := range c.Collecters {
		no := c.CheckAccess()
		if len(no) != 0 {
			m[c.Type()] = c.CheckAccess()
		}
	}
	if len(m) == 0 {
		return m, nil
	}
	if err := c.printNoAccessItem(m); err != nil {
		return m, err
	}
	return m, nil
}

func (c *CollecterHandler) printNoAccessItem(m map[string][]data.NoAccessRes) error {
	table := tabler.NewTable(
		"",
		tabler.NewRowTitle("TYPE", 0),
		tabler.NewRowTitle("COLLECT_ITEM", 0),
		tabler.NewRowTitle("DESCRIPTION", 0),
		tabler.NewRowTitle("TIPS", 0),
	)
	fmt.Printf("%s", bashdef.WithYellow("Detect some problem and some tips well give to you"))
	for t, noAccess := range m {
		for i, no := range noAccess {
			if i == 0 {
				err := table.AddColumn(t, no.ModuleItem, no.Description, no.Tips)
				if err != nil {
					log.Module.Errorf("add column err: %s", err.Error())
					return err
				}
				continue
			}
			if err := table.AddColumn("", no.ModuleItem, no.Description, no.Tips); err != nil {
				log.Module.Errorf("add column err: %s", err.Error())
				return err
			}
		}
	}
	fmt.Println()
	table.Print()
	var isConfirm string
	fmt.Printf("\nAre you want continue collect [y/n] ?\n")
	fmt.Scanln(&isConfirm)
	isConfirm = strings.ToLower(isConfirm)
	if isConfirm != "y" {
		fmt.Println("Stopping Collect.")
		return fmt.Errorf("some path permission denied, not continue collect")
	}
	return nil
}

func (c *CollecterHandler) printCollectItem(typeItem map[string][]string) error {
	table := tabler.NewTable(
		"",
		tabler.NewRowTitle("baseinfo", 0),
		tabler.NewRowTitle("diagnosis", 0),
		tabler.NewRowTitle("perfomance", 0),
	)
	moduleItems := [3][]string{}
	if base, ok := typeItem[collecttypedef.TYPE_BASE]; ok {
		moduleItems[0] = base
	}
	if diag, ok := typeItem[collecttypedef.TYPE_DIAG]; ok {
		moduleItems[1] = diag
	}
	if perf, ok := typeItem[collecttypedef.TYPE_PREF]; ok {
		moduleItems[2] = perf
	}
	for i := 0; i < len(moduleItems[0]) || i < len(moduleItems[1]) || i < len(moduleItems[2]); i++ {
		var (
			baseItem string
			diagItem string
			perfItem string
		)
		if i < len(moduleItems[0]) {
			baseItem = moduleItems[0][i]
		}
		if i < len(moduleItems[1]) {
			diagItem = moduleItems[1][i]
		}

		if i < len(moduleItems[2]) {
			perfItem = moduleItems[2][i]
		}
		if err := table.AddColumn(baseItem, diagItem, perfItem); err != nil {
			return nil
		}
	}
	fmt.Printf("%s\n", bashdef.WithBlue("The following modules will be collected"))
	table.Print()
	return nil
}

func (c *CollecterHandler) getCollectItem(noAccessMap map[string][]data.NoAccessRes) (typeItem map[string][]string, err error) {
	typeItem = make(map[string][]string)
	for _, collect := range c.Collecters {
		t := collect.Type()
		noAccess, ok := noAccessMap[t]
		if !ok {
			noAccess = make([]data.NoAccessRes, 0)
		}
		typeItem[t] = collect.CollectedItem(noAccess)
	}
	return
}

func (c *CollecterHandler) collect(moduleItems map[string][]string) error {
	progress := barutil.NewProgress()
	if e := c.Start(); e != nil {
		return e
	}
	collMap := c.collecterMap()
	for module, items := range moduleItems {
		_, ok := collMap[module]
		if !ok {
			log.Module.Errorf("collect type: %s not exist", module)
			continue
		}
		itemFunc := collMap[module].CollectFunc(items)
		progress.AddBar(module, itemFunc)
	}
	progress.Start()
	return c.Finsh()
}

func (c *CollecterHandler) Start() error {
	c.CollectResult.CollectBeginTime = time.Now()
	packageDir := c.CollectResult.GetPackageDir()
	for _, collecter := range c.Collecters {
		if err := collecter.Start(packageDir); err != nil {
			return err
		}
	}
	return nil
}

func (c *CollecterHandler) collecterMap() (res map[string]ytccollect.TypedCollecter) {
	res = make(map[string]ytccollect.TypedCollecter)
	for _, c := range c.Collecters {
		res[c.Type()] = c
	}
	return
}

func (c *CollecterHandler) Finsh() error {
	c.CollectResult.CollectEndtime = time.Now()
	for _, collecter := range c.Collecters {
		c.CollectResult.ModuleResults[collecter.Type()] = collecter.Finish()
	}
	path, err := c.CollectResult.GenResult(c.CollectResult.CollectParam.Output, data.TXT_REPORT, c.Types)
	if err != nil {
		log.Handler.Error(err)
		fmt.Printf("failed to gen result, err: %v\n", err)
		return err
	}
	fmt.Printf("result was saved to %s\n", path)
	return nil
}
