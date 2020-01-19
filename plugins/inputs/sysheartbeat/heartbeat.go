package sysheartbeat

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//SysHeartbeat config and other data type holder
type SysHeartbeat struct {
	AllPidsURL  string   `toml:"get_all_pid_url"`
	GetByPidURL string   `toml:"get_by_pid_url"`
	PidFilters  []string `toml:"pid_filters"`
}

var sampleConfig = `
  ## Sysctl input configuration
	##
	##Discover all the pids running in a system
  # get_all_pid_url = "http://localhost:7782/vm"
	##
	## URL to reterieve pids
  # get_by_pid_url = "http://localhost:7782/vm/{id}/stats"
	##
	## VM pids to be included / empty list means all pids
	# pid_filters = [ "eclipse" , "gradle", "tomcat"]
`

//Description of plugin values
func (cfg *SysHeartbeat) Description() string {
	return "sysctl heartbeat plugin"
}

//SampleConfig configuration
func (cfg *SysHeartbeat) SampleConfig() string {
	return sampleConfig
}

//Gather all the discovery field
func (cfg *SysHeartbeat) Gather(acc telegraf.Accumulator) error {

	//validate inputs
	if !cfg.validate() {
		err := fmt.Errorf("Sysctl hearbeat plugin does not have config 'AllPidsURL' %s or 'GetByPidURL' %s ", cfg.AllPidsURL, cfg.GetByPidURL)
		log.Println(err)
		acc.AddError(err)
		return nil
	}

	//get all pids from the service
	getAllPidResp, err := GetAllPids(cfg)
	if err != nil {
		err1 := fmt.Errorf("Error occured calling GetAllPids %v ", err)
		log.Println(err1)
		acc.AddError(err1)
		return nil
	}

	//Filter pid whose name matches filter
	var filteredPids *[]int

	if len(cfg.PidFilters) > 0 {

		filteredPids = cfg.getFilteredPids(&getAllPidResp)

	} else {

		allPids := make([]int, 0, 0)
		for _, pid := range getAllPidResp.Vms {
			allPids = append(allPids, pid.ID)
		}

		filteredPids = &allPids
	}

	log.Printf("Getting heartbeat for following pids after filteration %v \n", &filteredPids)

	getByPidRespMap, errArr := GetPidStats(cfg, *filteredPids)
	if len(errArr) > 0 {
		for _, err := range errArr {
			log.Printf("Error in get vm stats call %v \n", err)
		}
		return nil
	}

	for k, v := range getByPidRespMap {
		name := v.Name
		pid := k
		tags := map[string]string{
			"pid": strconv.Itoa(pid),
		}
		fields := map[string]interface{}{
			"name":                           name,
			"pid":                            strconv.Itoa(pid),
			"up_time":                        v.UpTime,
			"application_time":               v.ApplicationTime,
			"os_frequency":                   v.OsFrequency,
			"classes_loaded_classes":         v.Classes.LoadedClasses,
			"classes_unloaded_classes":       v.Classes.UnloadedClasses,
			"classes_sharedloaded_classes":   v.Classes.SharedLoadedClasses,
			"classes_sharedunloaded_classes": v.Classes.SharedUnloadedClasses,
			"threads_threads_live":           v.Threads.ThreadsLive,
			"threads_threads_daemon":         v.Threads.ThreadsDaemon,
			"threads_threads_livepeak":       v.Threads.ThreadsLivePeak,
			"threads_threads_started":        v.Threads.ThreadsStarted,
			"memory_genname":                 v.Memory.GenName,
			"memory_gencapacity":             v.Memory.GenCapacity,
			"memory_genused":                 v.Memory.GenUsed,
			"memory_gen_maxcapacity":         v.Memory.GenMaxCapacity,
			"additionalinfo":                 v.AdditionalInfo,
		}

		now := time.Now()

		acc.AddFields("sysheartbeat", fields, tags, now)

	}

	return nil
}

func (cfg *SysHeartbeat) getFilteredPids(getAllPidResp *GetAllPidResp) *[]int {
	filteredPids := make([]int, 0, 0)
	for _, filter := range cfg.PidFilters {
		filter = strings.ToLower(strings.Trim(filter, " "))
		if filter == "" {
			continue
		}

		for _, pid := range getAllPidResp.Vms {
			pidName := strings.ToLower(strings.Trim(pid.Name, " "))
			if pidName == "" {
				continue
			}
			if strings.Contains(pidName, filter) {
				filteredPids = append(filteredPids, pid.ID)
			}
		}
	}
	return &filteredPids
}

func (cfg *SysHeartbeat) validate() bool {
	if cfg.AllPidsURL == "" || cfg.GetByPidURL == "" {
		log.Printf("Sysctl hearbeat plugin does not have config 'AllPidsURL' %s or 'GetByPidURL' %s \n", cfg.AllPidsURL, cfg.GetByPidURL)
		return false
	}
	return true
}

func (cfg *SysHeartbeat) pidFiltersSet() map[string]int {
	filterMap := make(map[string]int)
	for _, pid := range cfg.PidFilters {
		filterMap[pid] = 1
	}
	return filterMap
}

func init() {
	inputs.Add("sysheartbeat", func() telegraf.Input { return &SysHeartbeat{} })
}
