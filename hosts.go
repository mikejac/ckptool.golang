/*
 * Copyright (c) 2016 Michael Jacobsen (github.com/mikejac)
 *
 * This file is part of ckptool.golang.
 *
 * ckptool.golang is free software: you can redistribute
 * it and/or modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * ckptool.golang is distributed in the hope that it will
 * be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
 * of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ckptool.golang.  If not,
 * see <http://www.gnu.org/licenses/>.
 *
 */

//
// https://github.com/go-ini/ini
//

package main

import (
	"strings"
	"github.com/go-ini/ini"
)

type HostsData struct {
	cfg	*ini.File
}

//
//
func NewHosts(hostsFile string) (hosts *HostsData, err error) {
	hosts = &HostsData{}
	
	hosts.cfg, err = ini.Load(hostsFile)
	if err != nil {
		return nil, err
	}
	
	return hosts, nil
}

//
//
func (hosts *HostsData) GetHostIP(host string) (ip string) {
	if hosts.cfg == nil {
		return host
	}
	
	names := hosts.cfg.SectionStrings()
	
	// let's see if we can find the name
	for _, n := range names {
		if hosts.cfg.Section(n).HasKey(host) {
			val := hosts.cfg.Section(n).Key(host).String()

			//fmt.Printf("HostsData.GetHost(): val = %s\n", val)
			
			v := strings.Split(val, ",")
			
			//fmt.Printf("HostsData.GetHost(): len = %d\n", len(v))
			
			for _, vv := range v {
				i := strings.Split(vv, ":")
				
				//fmt.Printf("HostsData.GetHost(): len = %d, i = %q\n", len(i), i)
				
				if len(v) == 1 && len(i) == 1 {							// host=192.168.1.1
					return i[0]
				} else if len(i) == 2 && i[0] == "ip" {
					return i[1]												// host=ip:192.168.1.1
				}
			}
		}
	}
	
	// didn't find the name
	return host
}

//
//
func (hosts *HostsData) GetClusterMembers(clusterName string) (members []string) {
	names := hosts.cfg.Section("cluster." + clusterName).KeyStrings()

	for _, n := range names {
		if !hosts.reservedKey(n) {
			members = append(members, n)
		}
	}
	
	return members
}

//
//
func (hosts *HostsData) GetClusterIgnoredRoutes(clusterName string) (routes map[string]struct{}) {
	routes = make(map[string]struct{}, 10)
	
	if hosts.cfg.Section("cluster." + clusterName).HasKey("ignore_routes") {
		val := hosts.cfg.Section("cluster." + clusterName).Key("ignore_routes").String()
	
		v := strings.Split(val, ",")
		
		for _, vv := range v {
			routes[vv] = struct{}{}
		}
	}
	
	return routes
}

//
//
func (hosts *HostsData) GetAllHosts() (h []string) {
	sections := hosts.cfg.SectionStrings()
	
	for _, s := range sections {
		names := hosts.cfg.Section(s).KeyStrings()
		
		for _, n := range names {
			if !hosts.reservedKey(n) {
				h = append(h, n)
			}
		}
	}
	
	return h
}

//
//
func (hosts *HostsData) GetAllStandalone() (h []string) {
	sections := hosts.cfg.SectionStrings()
	
	for _, s := range sections {
		if !strings.HasPrefix(s, "cluster.") {
			names := hosts.cfg.Section(s).KeyStrings()
			
			for _, n := range names {
				if !hosts.reservedKey(n) {
					h = append(h, n)
				}
			}
		}
	}

	return h
}

//
//
func (hosts *HostsData) GetAllCluster() (h []string) {
	sections := hosts.cfg.SectionStrings()
	
	for _, s := range sections {
		if strings.HasPrefix(s, "cluster.") {
			h = append(h, strings.TrimPrefix(s, "cluster."))
		}
	}

	return h
}
//
//
func (hosts *HostsData) reservedKey(key string) (yes bool) {
	if key == "ignore_routes" {
		return true
	}
	
	return false
}
