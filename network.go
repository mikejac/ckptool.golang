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
	"github.com/mikejac/ssh.golang"
)

//
//
func CompareNetworks(routes1 sshtool.Routes, routes2 sshtool.Routes, verbose int) (sharedRoutes sshtool.Routes, host1OnlyRoutes sshtool.Routes, host2OnlyRoutes sshtool.Routes) {
	if verbose >= 1 { fmt.Printf("CompareNetworks(): host1 routes in host2\n") }
	
	for _, r1 := range routes1 {
		if findNetwork(r1, routes2) {
			sharedRoutes = append(sharedRoutes, r1)
		} else {
			host1OnlyRoutes = append(host1OnlyRoutes, r1)			
		}
	}

	if verbose >= 1 { fmt.Printf("CompareNetworks(): host2 routes in host1\n") }
	
	for _, r2 := range routes2 {
		if findNetwork(r2, routes1) {
			
		} else {
			host2OnlyRoutes = append(host2OnlyRoutes, r2)			
		}
	}
	
	if verbose >= 1 {
		fmt.Printf("CompareNetworks(): sharedRoutes:\n")
		fmt.Printf("%q\n", sharedRoutes)
		fmt.Println()
		fmt.Printf("CompareNetworks(): host1OnlyRoutes:\n")
		fmt.Printf("%q\n", host1OnlyRoutes)
		fmt.Println()
		fmt.Printf("CompareNetworks(): host2OnlyRoutes:\n")
		fmt.Printf("%q\n", host2OnlyRoutes)
		fmt.Println()
	}
	
	return sharedRoutes, host1OnlyRoutes, host2OnlyRoutes
}

//
//
func findNetwork(n sshtool.NetworkRoute, routes sshtool.Routes) (found bool) {
	for _, r := range routes {
		if r.IPNet.IP.Equal(n.IPNet.IP) && r.IPNet.Mask.String() == n.IPNet.Mask.String() && r.Gateway == n.Gateway {
			if verbose >= 1 { fmt.Printf("findNetwork(): found; %s / %s -> %s\n", n.IPNet.IP.String(), n.IPNet.Mask.String(), n.Gateway) }
			
			return true
		}
	}

	if verbose >= 1 { fmt.Printf("findNetwork(): NOT found; %s / %s -> %s\n", n.IPNet.IP.String(), n.IPNet.Mask.String(), n.Gateway) }
	
	return false
}