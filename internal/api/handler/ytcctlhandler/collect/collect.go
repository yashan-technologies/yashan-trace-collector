package ytcctlhandler

import (
	"fmt"
	"sort"
	"strconv"
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

func (c *CollecterHandler) Collect(yasdbValidate error) error {
	noAccessMap, err := c.checkAccess(yasdbValidate)
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
	fmt.Printf("\nStarting collect...\n\n")
	return c.collect(moduleItems)
}

func (c *CollecterHandler) checkAccess(yasdbValidateErr error) (map[string][]data.NoAccessRes, error) {
	m := make(map[string][]data.NoAccessRes)
	for _, c := range c.Collecters {
		noAccessList := c.CheckAccess(yasdbValidateErr)
		if len(noAccessList) != 0 {
			m[c.Type()] = noAccessList
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
		tabler.NewRowTitle("TYPE", 15),
		tabler.NewRowTitle("COLLECT_ITEM", 25),
		tabler.NewRowTitle("DESCRIPTION", 50),
		tabler.NewRowTitle("TIPS", 50),
		tabler.NewRowTitle("COLLECTED?", 15),
	)
	fmt.Printf("%s\n\n", bashdef.WithYellow("Detect some problem and some tips will give to you"))
	var modules []string
	for t := range m {
		modules = append(modules, t)
	}
	sort.Strings(modules)
	for _, module := range modules {
		for i, noAccess := range m[module] {
			if i == 0 {
				err := table.AddColumn(strings.ToUpper(collecttypedef.GetTypeFullName(module)), noAccess.ModuleItem, noAccess.Description, noAccess.Tips, isColltedStr(noAccess.ForceCollect))
				if err != nil {
					log.Handler.Errorf("add column err: %s", err.Error())
					return err
				}
				continue
			}
			if err := table.AddColumn("", noAccess.ModuleItem, noAccess.Description, noAccess.Tips, isColltedStr(noAccess.ForceCollect)); err != nil {
				log.Module.Errorf("add column err: %s", err.Error())
				return err
			}
		}
	}
	table.Print()
	var isConfirm string
	fmt.Printf("\nAre you want continue collect [y/n] ?\n")
	fmt.Scanln(&isConfirm)
	isConfirm = strings.ToLower(isConfirm)
	if isConfirm != "y" {
		return fmt.Errorf("some validations failed, not continue collect")
	}
	return nil
}

func (c *CollecterHandler) printCollectItem(typeItem map[string][]string) error {
	var (
		itemTitle   []*tabler.RowTitle
		moduleItems = make([][]string, 0)
		moduleNames = make([]string, 0)
	)
	for module := range typeItem {
		moduleNames = append(moduleNames, module)
	}
	sort.Strings(moduleNames)
	for _, t := range moduleNames {
		itemTitle = append(itemTitle, tabler.NewRowTitle(strings.ToUpper(collecttypedef.GetTypeFullName(t)), 30))
		moduleItems = append(moduleItems, typeItem[t])
	}
	table := tabler.NewTable("", itemTitle...)
	maxCol := maxCol(moduleItems)
	for i := 0; i < maxCol; i++ {
		row := make([]interface{}, len(moduleNames))
		for j, item := range moduleItems {
			if i < len(item) {
				row[j] = item[i]
				continue
			}
			row[j] = " "
		}
		if err := table.AddColumn(row...); err != nil {
			return err
		}
	}
	fmt.Printf("%s\n\n", bashdef.WithBlue("The following modules will be collected"))
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
	progress := barutil.NewProgress(barutil.WithWidth(100))
	if e := c.Start(); e != nil {
		return e
	}
	collMap := c.collecterMap()
	for module, items := range moduleItems {
		_, ok := collMap[module]
		if !ok {
			log.Handler.Errorf("collect type: %s not exist", module)
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
	fmt.Printf("The collection has been %s and the result was saved to %s, thanks for your use.\n", bashdef.WithGreen("completed"), bashdef.WithBlue(path))
	return nil
}

func isColltedStr(f bool) string {
	flag := strconv.FormatBool(f)
	if f {
		return bashdef.WithGreen(flag)
	}
	return bashdef.WithRed(flag)
}

func maxCol(rows [][]string) int {
	max := -1
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}
