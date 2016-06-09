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

package main

import (
	"fmt"
	"strings"
	"io"
	"github.com/mikejac/ssh.golang"
)

type PrintData struct {
	writer io.Writer
}

//
//
func NewPrint(writer io.Writer) (print *PrintData) {
	print = &PrintData{
		writer: writer,
	}
	
	return print
}
//
//
func (print *PrintData) PrintInterfaces(physical sshtool.PhysicalInterfaces, logical sshtool.LogicalInterfaces) {
	var used map[string]bool
	used = make(map[string]bool, 0)

	print.writer.Write([]byte("# physical interfaces\n"))	
	
	// set interface eth1 state on
	for _, i := range physical {
		if _, ok := used[i.IfName]; !ok {
			print.writer.Write([]byte("set interface "))
			print.writer.Write([]byte(i.IfName))
			print.writer.Write([]byte(" state on\n"))		
			
			used[i.IfName] = true
		}			
	}

	print.writer.Write([]byte("# VLANs\n"))

	// add interface eth1 vlan 111
	for _, i := range physical {
		if i.VLAN != "" {
			print.writer.Write([]byte("add interface "))
			print.writer.Write([]byte(i.IfName))
			print.writer.Write([]byte(" vlan "))		
			print.writer.Write([]byte(i.VLAN))
			print.writer.Write([]byte("\n"))
		}			
	}

	print.writer.Write([]byte("# logical interfaces\n"))

	// set interface eth1.111 ipv4-address 192.168.1.1 mask-length 24
	for _, i := range logical {
		p := strings.Split(i.IfIP, "/")
		
		if len(p) > 1 {
			print.writer.Write([]byte("set interface "))
			print.writer.Write([]byte(i.IfName))
			print.writer.Write([]byte(" ipv4-address "))
			print.writer.Write([]byte(p[0]))
			print.writer.Write([]byte(" mask-length "))
			
			if len(p) == 2 {
				print.writer.Write([]byte(p[1]))
			} else {
				print.writer.Write([]byte("32"))
			}
			
			print.writer.Write([]byte("\n"))
		} else {
			print.writer.Write([]byte("# invalid ip/netmask\n"))
		}
	}
}

//
//
func (print *PrintData) PrintRoutes(routes sshtool.Routes) {
	print.writer.Write([]byte("# static routes\n"))	

	// set static-route 192.168.2.0/24 nexthop gateway address 192.168.1.10 priority 1 on
	for _, r := range routes {
		print.writer.Write([]byte("set static-route "))
		
		if r.IPNet.IP.String() == "0.0.0.0" {
			print.writer.Write([]byte("default"))			
		} else {		
			print.writer.Write([]byte(r.Net))
		}
		
		print.writer.Write([]byte(" nexthop gateway address "))
		print.writer.Write([]byte(r.Gateway))
		print.writer.Write([]byte(" priority 1 on\n"))
	}
}

//
//
func (print *PrintData) PrintCPHA(cpha *sshtool.CphaData) {
	print.writer.Write([]byte("# cpha state: "))
	print.writer.Write([]byte(cpha.Status))
	print.writer.Write([]byte("\n"))
}

//
//
func (print *PrintData) PrintComparedRoutes(sharedRoutes sshtool.Routes, host1OnlyRoutes sshtool.Routes, host2OnlyRoutes sshtool.Routes, ignoredRoutes map[string]struct{}) {
	fmt.Fprintf(print.writer, "Shared Routes (%d)\n", len(sharedRoutes))
	fmt.Fprintln(print.writer, "========================================================")
	
	if len(sharedRoutes) == 0 {
		fmt.Fprintln(print.writer, "(none)")
	} else {
		for _, r := range sharedRoutes {
			fmt.Fprintf(print.writer, "%-20s -> %-16s dev %s\n", r.Net, r.Gateway, r.Dev)
		}
	}

	fmt.Fprintln(print.writer)
	fmt.Fprintf(print.writer, "Host 1 Routes (%d)\n", len(host1OnlyRoutes))
	fmt.Fprintln(print.writer, "========================================================")
	
	if len(host1OnlyRoutes) == 0 {
		fmt.Fprintln(print.writer, "(none)")
	} else {
		for _, r := range host1OnlyRoutes {
			if _, ok := ignoredRoutes[r.Net]; !ok {
				fmt.Fprintf(print.writer, "%-20s -> %-16s dev %s\n", r.Net, r.Gateway, r.Dev)
			} else {
				fmt.Fprintf(print.writer, "Ignored: %-20s -> %-16s\n", r.Net, r.Gateway)				
			}
		}
	}

	fmt.Fprintln(print.writer)
	fmt.Fprintf(print.writer, "Host 2 Routes (%d)\n", len(host2OnlyRoutes))
	fmt.Fprintln(print.writer, "========================================================")
	
	if len(host2OnlyRoutes) == 0 {
		fmt.Fprintln(print.writer, "(none)")
	} else {
		for _, r := range host2OnlyRoutes {
			if _, ok := ignoredRoutes[r.Net]; !ok {
				fmt.Fprintf(print.writer, "%-20s -> %-16s dev %s\n", r.Net, r.Gateway, r.Dev)
			} else {
				fmt.Fprintf(print.writer, "Ignored: %-20s -> %-16s\n", r.Net, r.Gateway)				
			}
		}
	}
}