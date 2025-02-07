// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// CPU information.
type CPU struct {
	Vendor  string `json:"vendor,omitempty"`
	Model   string `json:"model,omitempty"`
	Speed   uint   `json:"speed,omitempty"`   // CPU clock rate in MHz
	Cache   uint   `json:"cache,omitempty"`   // CPU cache size in KB
	Cpus    uint   `json:"cpus,omitempty"`    // number of physical CPUs
	Cores   uint   `json:"cores,omitempty"`   // number of physical CPU cores
	Threads uint   `json:"threads,omitempty"` // number of logical (HT) CPU cores
}

var (
	reTwoColumns = regexp.MustCompile("\t+: ")
	reExtraSpace = regexp.MustCompile(" +")
	reCacheSize  = regexp.MustCompile(`^(\d+) KB$`)
)

func GetCPUInfo(virtualEnv bool) CPU {
	cpuInfo := CPU{}
	cpuInfo.Threads = uint(runtime.NumCPU())

	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return cpuInfo
	}
	defer f.Close()

	cpu := make(map[string]bool)
	core := make(map[string]bool)

	var cpuID string

	s := bufio.NewScanner(f)
	for s.Scan() {
		if sl := reTwoColumns.Split(s.Text(), 2); sl != nil {
			switch sl[0] {
			case "physical id":
				cpuID = sl[1]
				cpu[cpuID] = true
			case "core id":
				coreID := fmt.Sprintf("%s/%s", cpuID, sl[1])
				core[coreID] = true
			case "vendor_id":
				if cpuInfo.Vendor == "" {
					cpuInfo.Vendor = sl[1]
				}
			case "model name":
				if cpuInfo.Model == "" {
					// CPU model, as reported by /proc/cpuinfo, can be a bit ugly. Clean up...
					model := reExtraSpace.ReplaceAllLiteralString(sl[1], " ")
					cpuInfo.Model = strings.Replace(model, "- ", "-", 1)
				}
			case "cache size":
				if cpuInfo.Cache == 0 {
					if m := reCacheSize.FindStringSubmatch(sl[1]); m != nil {
						if cache, err := strconv.ParseUint(m[1], 10, 64); err == nil {
							cpuInfo.Cache = uint(cache)
						}
					}
				}
			}
		}
	}
	if s.Err() != nil {
		return cpuInfo
	}

	if !virtualEnv {
		cpuInfo.Cpus = uint(len(cpu))
		cpuInfo.Cores = uint(len(core))
	}
	return cpuInfo
}
