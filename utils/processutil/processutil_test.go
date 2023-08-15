package processutil_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"ytc/utils/processutil"
)

func TestIsRunning(t *testing.T) {
	p := processutil.NewProcess(1)
	_, ok := p.IsRunning()
	if !ok {
		t.Fail()
	}
}

func TestListProcess(t *testing.T) {
	processes, err := processutil.ListProcess()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.MarshalIndent(processes, "", "  ")
	t.Log(len(processes))
	t.Log(string(data))
}

func TestListProcessByCmdline(t *testing.T) {
	ps := []string{"/home/lx/yashandb/ha_home/node_1", "/home/lx/yashandb/ha_home/node_2", "/home/lx/yashandb/ha_home/node_3"}
	for _, p := range ps {
		fmt.Println(p)
		match := fmt.Sprintf(".*yasdb (?i:(nomount|mount|open)) -D %s", p)
		processes, err := processutil.ListProcessByCmdline("lx", match, true)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(processes)
	}
}
