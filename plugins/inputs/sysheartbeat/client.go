package sysheartbeat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//GetAllPidResp response for Endpoint http://localhost:7782/vm
type GetAllPidResp struct {
	Vms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Desc string `json:"desc"`
	} `json:"vms"`
}

//GetByPidResp for endpoint http://localhost:7782/vm/33760/stats
type GetByPidResp struct {
	Name            string `json:"name"`
	UpTime          int    `json:"upTime"`
	ApplicationTime int    `json:"applicationTime"`
	OsFrequency     int    `json:"osFrequency"`
	Classes         struct {
		LoadedClasses         int `json:"loadedClasses"`
		UnloadedClasses       int `json:"unloadedClasses"`
		SharedLoadedClasses   int `json:"sharedLoadedClasses"`
		SharedUnloadedClasses int `json:"sharedUnloadedClasses"`
	} `json:"classes"`
	Threads struct {
		ThreadsLive     int `json:"threadsLive"`
		ThreadsDaemon   int `json:"threadsDaemon"`
		ThreadsLivePeak int `json:"threadsLivePeak"`
		ThreadsStarted  int `json:"threadsStarted"`
	} `json:"threads"`
	Memory struct {
		GenName        []string `json:"genName"`
		GenCapacity    []int    `json:"genCapacity"`
		GenUsed        []int    `json:"genUsed"`
		GenMaxCapacity []int    `json:"genMaxCapacity"`
	} `json:"memory"`
	AdditionalInfo interface{} `json:"additionalInfo"`
}

//GetPidStats of vm endpoint
// improvements use channels and go routines to make it faster
func GetPidStats(config *SysHeartbeat, pids []int) (map[int]GetByPidResp, []error) {

	pidMap := make(map[int]GetByPidResp)
	errorSlice := make([]error, 0, len(pids))

	if config.GetByPidURL == "" {
		return pidMap, append(errorSlice, errors.New("url is nil in sysctl GetPidStats"))
	}

	if len(pids) == 0 {
		return pidMap, append(errorSlice, errors.New("url is nil in sysctl GetPidStats"))
	}

	fmt.Printf("looping over %v \n", pids)
	for _, pid := range pids {

		fmt.Printf("iteration for pid %d \n", pid)
		pidURL := strings.Replace(config.GetByPidURL, "{id}", strconv.Itoa(pid), -1)
		_, errURL := url.Parse(pidURL)

		if errURL != nil {
			log.Printf("Malformed url %s,\n %v \n", pidURL, errURL)
			errorSlice = append(errorSlice, fmt.Errorf("malformed url %s, %v ", pidURL, errURL))
			continue
		}
		fmt.Printf("calling url %s \n", pidURL)
		resp, err := http.Get(pidURL)
		if err != nil {
			log.Printf("Error occured when calling url %s \n %v \n", pidURL, err)
			errorSlice = append(errorSlice, fmt.Errorf("Error occured when calling url %s, %v", pidURL, err))
			continue
		}

		defer resp.Body.Close()

		var pidResp GetByPidResp

		if err := json.NewDecoder(resp.Body).Decode(&pidResp); err != nil {
			log.Fatal("Error in calling pid stats service ", err)
			errorSlice = append(errorSlice, fmt.Errorf("Error in calling pid stats service %v", err))
			continue
		}

		pidMap[pid] = pidResp

	}

	return pidMap, nil
}

//GetAllPids from sysctl jvm agent
func GetAllPids(config *SysHeartbeat) (GetAllPidResp, error) {
	url := config.AllPidsURL
	var allPidResp GetAllPidResp

	if url == "" {
		return allPidResp, errors.New("url is nil in GetAllPids")
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error in calling GetAllPids service in sysctl heartbeat plugin", err)
		return allPidResp, err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&allPidResp); err != nil {
		log.Fatal("Error in calling GetAllPids service in sysctl heartbeat plugin", err)
		return allPidResp, err
	}

	return allPidResp, nil
}
