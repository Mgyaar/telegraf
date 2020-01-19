package sysheartbeat

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	cfg := &SysHeartbeat{
		AllPidsURL:  "http://localhost:7782/vm",
		GetByPidURL: "http://localhost:7782/vm/{id}/stats",
		PidFilters:  []string{"jvmagent", "gradle", "tomcat", "eclipse"},
	}
	acc := &testutil.Accumulator{}

	cfg.Gather(acc)
}
