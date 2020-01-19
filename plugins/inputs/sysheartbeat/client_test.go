package sysheartbeat

import (
	"fmt"
	"testing"
)

func TestGetVMIds(t *testing.T) {
	config := &SysHeartbeat{
		AllPidsURL: "http://localhost:7782/vm",
	}

	fmt.Println(config)

	resp, err := GetAllPids(config)
	if err != nil {
		t.Error("Error reported ", err)
	}

	for _, vm := range resp.Vms {
		fmt.Println(vm)
	}
}

func TestGetVMStats(t *testing.T) {
	config := &SysHeartbeat{
		AllPidsURL:  "http://localhost:7782/vm",
		GetByPidURL: "http://localhost:7782/vm/{id}/stats",
	}

	resp, err := GetAllPids(config)
	if err != nil {
		t.Error("Error reported ", err)
	}

	ids := make([]int, 0, len(resp.Vms))
	for _, vm := range resp.Vms {
		ids = append(ids, vm.ID)
	}
	fmt.Println(ids)
	maps, err2 := GetPidStats(config, ids)
	if err2 != nil {
		t.Error("Error reported ", err2)
	}
	fmt.Println(maps)
}
