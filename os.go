// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

package sysinfo

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// OS information.
type OS struct {
	Name         string `json:"name,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	Version      string `json:"version,omitempty"`
	Release      string `json:"release,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

var (
	rePrettyName = regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	reID         = regexp.MustCompile(`^ID=(.*)$`)
	reVersionID  = regexp.MustCompile(`^VERSION_ID=(.*)$`)
	reUbuntu     = regexp.MustCompile(`[\( ]([\d\.]+)`)
	reCentOS     = regexp.MustCompile(`^CentOS( Linux)? release ([\d\.]+) `)
	reRedHat     = regexp.MustCompile(`[\( ]([\d\.]+)`)
)

func GetOSInfo() OS {
	osInfo := OS{}
	// This seems to be the best and most portable way to detect OS architecture (NOT kernel!)
	if _, err := os.Stat("/lib64/ld-linux-x86-64.so.2"); err == nil {
		osInfo.Architecture = "amd64"
	} else if _, err := os.Stat("/lib/ld-linux.so.2"); err == nil {
		osInfo.Architecture = "i386"
	}

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return osInfo
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := rePrettyName.FindStringSubmatch(s.Text()); m != nil {
			osInfo.Name = strings.Trim(m[1], `"`)
		} else if m := reID.FindStringSubmatch(s.Text()); m != nil {
			osInfo.Vendor = strings.Trim(m[1], `"`)
		} else if m := reVersionID.FindStringSubmatch(s.Text()); m != nil {
			osInfo.Version = strings.Trim(m[1], `"`)
		}
	}

	switch osInfo.Vendor {
	case "debian":
		osInfo.Release = slurpFile("/etc/debian_version")
	case "ubuntu":
		if m := reUbuntu.FindStringSubmatch(osInfo.Name); m != nil {
			osInfo.Release = m[1]
		}
	case "centos":
		if release := slurpFile("/etc/centos-release"); release != "" {
			if m := reCentOS.FindStringSubmatch(release); m != nil {
				osInfo.Release = m[2]
			}
		}
	case "rhel":
		if release := slurpFile("/etc/redhat-release"); release != "" {
			if m := reRedHat.FindStringSubmatch(release); m != nil {
				osInfo.Release = m[1]
			}
		}
		if osInfo.Release == "" {
			if m := reRedHat.FindStringSubmatch(osInfo.Name); m != nil {
				osInfo.Release = m[1]
			}
		}
	}
	return osInfo
}
