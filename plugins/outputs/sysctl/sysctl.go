package sysctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/shirou/gopsutil/host"
	gopnet "github.com/shirou/gopsutil/net"
)

var sampleConfig = `
## Sysctl to write to, "stdout" is a specially handled file.
url_path = "http://localhost:8181/publishEvent"

## Data format to output.
## Each data format has it's own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
## data_format = "json"
building_block = "datacenter1"`

//Sysctl struct type
type Sysctl struct {
	URLPath       string `toml:"url_path"`
	BuildingBlock string `toml:"building_block"`
	Tenant        string `toml:"tenant"`
}

//CrudeEvent type
type CrudeEvent struct {
	TenantID                      string            `json:"tenantID"`
	EventID                       string            `json:"eventID"`
	EnterpriseMessageID           string            `json:"enterpriseMessageID"`
	EventDateTime                 string            `json:"eventDateTime"`
	ComponentID                   string            `json:"componentID"`
	ParentComponentID             string            `json:"parentComponentID"`
	LocalComponentID              string            `json:"localComponentID"`
	ExternalComponentID           string            `json:"externalComponentID"`
	HostName                      string            `json:"hostName"`
	ComponentStatus               string            `json:"componentStatus"`
	ComponentStatusDateChangeTime string            `json:"componentStatusDateChangeTime"`
	ApplicationName               string            `json:"applicationName"`
	NextApplicationName           string            `json:"nextApplicationName"`
	BuildingBlockName             string            `json:"buildingBlockName"`
	ComponentProperties           map[string]string `json:"componentProperties"`
}

const (
	debug = false
)

func getIP() string {
	var IP string
	statsIP, _ := gopnet.Interfaces()
	for _, val := range statsIP {
		if val.Name == "en0" {
			if len(val.Addrs) == 2 {
				IP = val.Addrs[1].Addr
			}
		}
	}
	if IP == "" {
		IP = "127.0.0.1"
	}
	return IP
}

//Connect struct
func (sys *Sysctl) Connect() error {
	return nil
}

// Close any connections to the Output
func (sys *Sysctl) Close() error {
	return nil
}

// Description returns a one-sentence description on the Output
func (sys *Sysctl) Description() string {
	return "Sysctl style output"
}

// SampleConfig returns the default configuration of the Output
func (sys *Sysctl) SampleConfig() string {
	return sampleConfig
}

// Write takes in group of points to be written to the Output
func (sys *Sysctl) Write(metrics []telegraf.Metric) error {

	if sys.URLPath == "" {
		panic("sysctl url should have a legal value")
	}

	propertiesMap := make(map[string]string)
	event := CrudeEvent{}
	hostName, _ := os.Hostname()
	event.ComponentID = hostName
	event.ComponentStatus = ""
	event.ComponentStatusDateChangeTime = ""
	event.EnterpriseMessageID = ""
	event.EventDateTime = time.Now().Format(time.RFC3339)
	event.EventID = ""
	event.ExternalComponentID = ""
	info, _ := host.Info()
	event.HostName = info.Hostname
	event.LocalComponentID = ""
	event.NextApplicationName = ""
	event.ParentComponentID = ""
	event.TenantID = sys.Tenant
	event.BuildingBlockName = sys.BuildingBlock

	for _, metric := range metrics {

		name := metric.Name()
		tag := fmt.Sprintf("%s", name)
		for k, v := range metric.Tags() {
			if k == "host" || strings.TrimSpace(k) == "" {
				continue
			}
			tag += "."
			tag += fmt.Sprintf("%s.%s", k, v)
		}

		for k, v := range metric.Fields() {
			field := tag + "."
			propertiesMap[fmt.Sprintf("%s%s", field, k)] = fmt.Sprintf("%v", v)
		}

	}

	propertiesMap[fmt.Sprintf("global.buildingblock")] = fmt.Sprintf("%v", sys.BuildingBlock)
	propertiesMap[fmt.Sprintf("global.ip")] = fmt.Sprintf("%v", getIP())

	event.ComponentProperties = propertiesMap

	body, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	req, err := http.NewRequest("POST", sys.URLPath, bytes.NewBuffer(body))

	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)

	return nil
}

func init() {
	outputs.Add("sysctl", func() telegraf.Output { return &Sysctl{} })
}
