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
	"os"
	"strings"
	"github.com/mikejac/ssh.golang"
	"github.com/docopt/docopt-go"
)

/*
	usage := `Naval Fate.

Usage:
  naval_fate ship new <name>...
  naval_fate ship <name> move <x> <y> [--speed=<kn>]
  naval_fate ship shoot <x> <y>
  naval_fate mine (set|remove) <x> <y> [--moored|--drifting]
  naval_fate -h | --help
  naval_fate --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --speed=<kn>  Speed in knots [default: 10].
  --moored      Moored (anchored) mine.
  --drifting    Drifting mine.`
*/

type HostData struct {
	Name					string

	Osclass				sshtool.OsClass
	Ostype					sshtool.OsType
	
	LogicalInterfaces		sshtool.LogicalInterfaces
	PhysicalInterfaces	sshtool.PhysicalInterfaces
	Routes					sshtool.Routes
	Cpha					*sshtool.CphaData
	
	//ConnectOk				bool
	ConnectText			string
	
	//InfoOk					bool
	//InfoText				string
	FwVer					string
	Platform				string
	
	Errors					uint
}

type ClusterData struct {
	Name					string
	Hosts					map[string]HostData
	Routes					map[string]sshtool.Routes
	
	Errors					uint
}

const (
	flagSummary				uint = 0x01
	
	errConnect					uint = 0x01
	errOS						uint = 0x02
	errLogicalInterfaces		uint = 0x04
	errPhysicalInterfaces	uint = 0x08
	errRoutes					uint = 0x10
	errCpha					uint = 0x20
	
	errRouteMismatch			uint = 0x01
	errCphaStat				uint = 0x02
)

var (
	verbose	int
	flags		uint
	hostsFile	string	= "hosts.ini"
)

func main() {
	usage := `Ckp Tool.

Usage:
  ckptool [--verbose] cluster host1 <host1> host2 <host2> user <username>
  ckptool [--verbose] cluster name <cluster-name> user <username>
  ckptool [--verbose] migrate host <host> user <username>
  ckptool [--verbose] xbm <host> user <username>
  ckptool [--verbose] check user <username> [--summary]
  ckptool [--verbose] all user <username>
  ckptool -h | --help
  ckptool --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --verbose     Verbose output.`

	arguments, _ := docopt.Parse(usage, nil, true, "Ckp Tool 1.0", false)
	
	if arguments["--verbose"].(bool) {
		verbose = 1
		fmt.Println(arguments)
		fmt.Println()
	}
	
	if arguments["--summary"].(bool) {
		flags += flagSummary
	}

	hosts, err := NewHosts(hostsFile)	
	if err != nil {
		fmt.Printf("WARNING: failed to load hostsfile (%s): %s\n", hostsFile, err.Error())		
	}
			
	print := NewPrint(os.Stdout)
	
	if arguments["xbm"].(bool) {
		host := hosts.GetHostIP(arguments["<host>"].(string))
		
		password, ok := Credentials("SSH Password: ")
		if !ok {
			return
		}

		expert_password, ok := Credentials("Unix Password: ")
		if !ok {
			return
		}
		
		fmt.Println("Host: " + host)
		
		doXBM(host, arguments["<username>"].(string), password, expert_password, 22, verbose)
		
	} else if arguments["cluster"].(bool) {
		var host1 string
		var host2 string
				
		if arguments["name"].(bool) {
			members := hosts.GetClusterMembers(arguments["<cluster-name>"].(string))

			if len(members) == 2 {
				host1 = hosts.GetHostIP(members[0])
				host2 = hosts.GetHostIP(members[1])
			} else {
				fmt.Printf("ERROR: cluster does not contain exactly two members\n")		
			}
		} else {
			host1 = hosts.GetHostIP(arguments["<host1>"].(string))
			host2 = hosts.GetHostIP(arguments["<host2>"].(string))
		}

		password, ok := Credentials("SSH Password: ")
		if !ok {
			return
		}

		expert_password, ok := Credentials("Expert Password: ")
		if !ok {
			return
		}

		fmt.Println("Host 1: " + host1)
		hostData1, ok1 := doHost(host1, arguments["<username>"].(string), password, expert_password, 22, verbose)
		
		print.PrintCPHA(hostData1.Cpha)
		fmt.Println()
		
		fmt.Println("Host 2: " + host2)
		hostData2, ok2 := doHost(host2, arguments["<username>"].(string), password, expert_password, 22, verbose)

		print.PrintCPHA(hostData2.Cpha)
		fmt.Println()
		
		if ok1 && ok2 {
			sharedRoutes, host1OnlyRoutes, host2OnlyRoutes := CompareNetworks(hostData1.Routes, hostData2.Routes, verbose)
			
			print.PrintComparedRoutes(sharedRoutes, host1OnlyRoutes, host2OnlyRoutes, hosts.GetClusterIgnoredRoutes(arguments["<cluster-name>"].(string)))
		}
	} else if arguments["migrate"].(bool) {
		host := hosts.GetHostIP(arguments["<host>"].(string))
		
		password, ok := Credentials("SSH Password: ")
		if !ok {
			return
		}

		expert_password, ok := Credentials("Expert Password: ")
		if !ok {
			return
		}
		
		fmt.Println("Host: " + host)
		
		if hostData, ok := doHost(host, arguments["<username>"].(string), password, expert_password, 22, verbose); ok {
			fmt.Println("# host: " + arguments["<host>"].(string))
			
			// now print the data
			print.PrintCPHA(hostData.Cpha)
			print.PrintInterfaces(hostData.PhysicalInterfaces, hostData.LogicalInterfaces)
			print.PrintRoutes(hostData.Routes)
		}
	} else if arguments["check"].(bool) {
		allStandalone := hosts.GetAllStandalone()
		allCluster    := hosts.GetAllCluster()
		
		password, ok := Credentials("SSH Password: ")
		if !ok {
			return
		}

		expert_password, ok := Credentials("Expert Password: ")
		if !ok {
			return
		}

		/******************************************************************************************************************
		 * extract data from all hosts and clusters
		 *
		 */
		
		var hostData []HostData
		hostData = make([]HostData, 0)
		
		for _, h := range allStandalone {
			hd, ok := checkStandalone(hosts, h, arguments["<username>"].(string), password, expert_password, 22, verbose)
			if !ok {
				hostData = append(hostData, hd)
			}
		}
		
		var clusterData []ClusterData
		clusterData = make([]ClusterData, 0)
		
		for _, c := range allCluster {
			cd, ok := checkCluster(hosts, c, arguments["<username>"].(string), password, expert_password, 22, flags, verbose)
			if !ok {
				clusterData = append(clusterData, cd)
			}
		}
		
		/******************************************************************************************************************
		 * print summary
		 *
		 */

		if (flags & flagSummary) != 0 {
			fmt.Println()
			fmt.Printf("=========================================================\n")
			fmt.Printf("Summary\n")
			fmt.Printf(" Number of hosts with issues ...: %d\n", len(hostData))
			fmt.Printf(" Number of clusters with issues : %d\n", len(clusterData))
			fmt.Println()

			fmt.Printf("Hosts\n")

			for _, h := range hostData {
				fmt.Printf("  Host: %s\n", h.Name /*name*/)
				
				if (h.Errors & errConnect) != 0 {
					fmt.Printf("   Error: could not connect to host\n")
				} else {
					if (h.Errors & errOS) != 0 {
						fmt.Printf("   Error: could not retrieve OS information\n")
					}
					if (h.Errors & errLogicalInterfaces) != 0 {
						fmt.Printf("   Error: could not retrieve logical interface\n")
					}
					if (h.Errors & errPhysicalInterfaces) != 0 {
						fmt.Printf("   Error: could not retrieve physical interface\n")
					}
					if (h.Errors & errRoutes) != 0 {
						fmt.Printf("   Error: could not retrieve routes\n")
					}
					if (h.Errors & errCpha) != 0 {
						fmt.Printf("   Error: could not retrieve CPHA information\n")
					}
				}

				fmt.Println()
			}
			
			fmt.Printf("Clusters\n")
			
			for _, c := range clusterData {
				fmt.Printf(" Cluster name: %s\n", c.Name)

				if (c.Errors & errRouteMismatch) != 0 {
					fmt.Printf("  Error: routes do not match on cluster members\n")
				}
				if (c.Errors & errCphaStat) != 0 {
					fmt.Printf("  Error: CPHA not working\n")
				}
				fmt.Println()
				
				for _/*name*/, h := range c.Hosts {
					fmt.Printf("  Host: %s\n", h.Name /*name*/)
					
					if (h.Errors & errConnect) != 0 {
						fmt.Printf("   Error: could not connect to host\n")
					} else {
						if (h.Errors & errOS) != 0 {
							fmt.Printf("   Error: could not retrieve OS information\n")
						}
						if (h.Errors & errLogicalInterfaces) != 0 {
							fmt.Printf("   Error: could not retrieve logical interface\n")
						}
						if (h.Errors & errPhysicalInterfaces) != 0 {
							fmt.Printf("   Error: could not retrieve physical interface\n")
						}
						if (h.Errors & errRoutes) != 0 {
							fmt.Printf("   Error: could not retrieve routes\n")
						}
						if (h.Errors & errCpha) != 0 {
							fmt.Printf("   Error: could not retrieve CPHA information\n")
						}
						if len(c.Routes[h.Name]) > 0 {
							fmt.Printf("   Mismatched routes:\n")
							
							for _, r := range c.Routes[h.Name] {
								fmt.Printf("    %-20s -> %-16s dev %s\n", r.Net, r.Gateway, r.Dev)
							}
							
							fmt.Println()
						}
					}
				}

				fmt.Println()
			}
		}
	} else if arguments["all"].(bool) {
		allHosts := hosts.GetAllHosts()
		
		password, ok := Credentials("SSH Password: ")
		if !ok {
			return
		}

		expert_password, ok := Credentials("Expert Password: ")
		if !ok {
			return
		}
		
		for _, h := range allHosts {
			fmt.Println("Host: " + h)
			
			doHost(hosts.GetHostIP(h), arguments["<username>"].(string), password, expert_password, 22, verbose)
			fmt.Println("========================================================")
		}
	}
}

//
//
func doHost(host string, user string, passw string, su_passw string, port int, verbose int) (hostData HostData, ok bool) {
	ssh, err := sshtool.NewSshAction(host, user, passw, su_passw, port, verbose)

	if err == nil {
		var logical	sshtool.LogicalInterfaces
		var physical	sshtool.PhysicalInterfaces
		var routes		sshtool.Routes
		var cpha		*sshtool.CphaData
		
		fmt.Printf("Connecting ... ")
		
		if err := ssh.Connect(); err == nil {
			fmt.Printf("done\n")
			fmt.Printf("Retriveing OS information ... ")

			if _, _, err := ssh.GetOS(); err == nil {
				fmt.Printf("done\n")
				fmt.Printf("Retrieving logical interface information ... ")

				if logical, err = ssh.GetInterfaces(); err == nil {
					fmt.Printf("done\n")
					fmt.Printf("Retrieving physical interface information ... ")

					if physical, err = ssh.GetPhyInterfaces(logical); err == nil {
						fmt.Printf("done\n")
						fmt.Printf("Retrieving routes ... ")

						if routes, err = ssh.GetRoutes(); err == nil {
							fmt.Printf("done\n")
							fmt.Printf("Retrieving HA information ... ")

							if cpha, err = ssh.GetCPHA(); err == nil {
								fmt.Printf("done\n\n")

								hostData.LogicalInterfaces	= logical
								hostData.PhysicalInterfaces	= physical
								hostData.Routes				= routes
								hostData.Cpha					= cpha
								
								return hostData, true
							} else {
								fmt.Println("error: " + err.Error())		
							}
						} else {
							fmt.Println("error: " + err.Error())		
						}
					} else {
						fmt.Println("error: " + err.Error())		
					}
				} else {
					fmt.Println("error: " + err.Error())		
				}
			} else {
				fmt.Println("error: " + err.Error())		
			}
			
			ssh.Exit()
			ssh.Disconnect()
		}
	} else {
		fmt.Println("error: " + err.Error())		
	}
	
	return hostData, false
}

//
//
func doXBM(host string, user string, passw string, su_passw string, port int, verbose int) (ok bool) {
	ssh, err := sshtool.NewSshAction(host, user, passw, su_passw, port, verbose)

	var osclass		sshtool.OsClass

	if err == nil {
		fmt.Printf("Connecting ... ")
		
		if err := ssh.Connect(); err == nil {
			fmt.Printf("done\n")
			fmt.Printf("Retriveing OS information ... ")

			if osclass, _, err = ssh.GetOS(); err == nil {
				fmt.Printf("done\n")
				
				if osclass == sshtool.OsClassXBM {
					if vapGroups, err := ssh.GetVAPGroups(); err == nil {
						fmt.Printf("vapGroups = %v\n", vapGroups)	
						
						if err = ssh.ConnectVAP("X02_RTVLPA_DK", 1); err == nil {
							
							ssh.DisconnectVAP()
						} else {
							fmt.Println("error: " + err.Error())								
						}
					} else {
						fmt.Println("error: " + err.Error())								
					}

				} else {
					fmt.Println("error: host is not an CrossBeam CPM")							
				}

			} else {
				fmt.Println("error: " + err.Error())		
			}
			
			ssh.Exit()
			ssh.Disconnect()
		}
	} else {
		fmt.Println("error: " + err.Error())		
	}
	
	return false
}

//
//
func checkStandalone(hosts *HostsData, hostname string, user string, passw string, su_passw string, port int, verbose int) (hostData HostData, ok bool) {
	hostData.Name = hostname
	
	h := hosts.GetHostIP(hostname)

	fmt.Printf("host:%s:addr:%s\n", hostname, h)
	
	ssh, err := sshtool.NewSshAction(h, user, passw, su_passw, port, verbose)

	if err == nil {
		if err := ssh.Connect(); err == nil {
			var fwver		string
			var platform	string
			var logical	sshtool.LogicalInterfaces
			var physical	sshtool.PhysicalInterfaces
			var routes		sshtool.Routes
			var cpha		*sshtool.CphaData
			
			if hostData.Osclass, hostData.Ostype, err = ssh.GetOS(); err == nil {
				if fwver, platform, err = ssh.GetInfo(); err == nil {
					hostData.FwVer	= fwver
					hostData.Platform	= platform
					
					fmt.Printf("host:%s:fwver:\"%s\"\n", hostname, fwver)
					fmt.Printf("host:%s:platform:\"%s\"\n", hostname, platform)
				} else {
					hostData.Errors |= errOS

					fmt.Printf("host:%s:fwver:null\n", hostname)
					fmt.Printf("host:%s:platform:null\n", hostname)
				}
				
				if logical, err = ssh.GetInterfaces(); err == nil {
					if len(logical) == 0 {
						hostData.Errors |= errLogicalInterfaces
						fmt.Printf("host:%s:logical:false\n", hostname)
					} else {
						hostData.LogicalInterfaces = logical
						fmt.Printf("host:%s:logical:true\n", hostname)
					}
				} else {
					hostData.Errors |= errLogicalInterfaces
					fmt.Printf("host:%s:logical:false\n", hostname)
				}
				
				if physical, err = ssh.GetPhyInterfaces(logical); err == nil {
					if len(physical) == 0 {
						hostData.Errors |= errPhysicalInterfaces
						fmt.Printf("host:%s:physical:false\n", hostname)
					} else {
						hostData.PhysicalInterfaces = physical
						fmt.Printf("host:%s:physical:true\n", hostname)
					}
				} else {
					hostData.Errors |= errPhysicalInterfaces
					fmt.Printf("host:%s:physical:false\n", hostname)
				}
				
				if routes, err = ssh.GetRoutes(); err == nil {
					if len(routes) == 0 {
						hostData.Errors |= errRoutes
						fmt.Printf("host:%s:routes:false\n", hostname)
					} else {
						hostData.Routes = routes
						fmt.Printf("host:%s:routes:true\n", hostname)
					}
				} else {
					hostData.Errors |= errRoutes
					fmt.Printf("host:%s:routes:false\n", hostname)
				}
				
				if cpha, err = ssh.GetCPHA(); err == nil {
					hostData.Cpha = cpha
					fmt.Printf("host:%s:cpha:true\n", hostname)
				} else {
					hostData.Errors |= errCpha
					fmt.Printf("host:%s:cpha:false\n", hostname)
				}
			} else {
				hostData.Errors |= errOS
			}
			
			ssh.Exit()
			ssh.Disconnect()
		} else {
			hostData.ConnectText = err.Error()
			hostData.Errors |= errConnect
		}
	} else {
		hostData.Errors |= errConnect
		hostData.ConnectText	= err.Error()
	}

	if hostData.Errors == 0 {
		fmt.Printf("host:%s:ok:true\n", hostname)
		ok = true
	} else {
		fmt.Printf("host:%s:ok:false\n", hostname)
		ok = false
	}
	
	return hostData, ok
}

//
//
func checkCluster(hosts *HostsData, clustername string, user string, passw string, su_passw string, port int, flags uint, verbose int) (clusterData ClusterData, ok bool) {
	clusterData.Hosts		= make(map[string]HostData)
	clusterData.Routes	= make(map[string]sshtool.Routes)
	clusterData.Name		= clustername
	ok						= true
	
	members := hosts.GetClusterMembers(clustername)

	if len(members) == 2 {
		hostData1, ok1 := checkStandalone(hosts, members[0], user, passw, su_passw, 22, verbose)
		
		clusterData.Hosts[members[0]] = hostData1

		if ok1 {
			fmt.Printf("host:%s:cpha:\"%s\"\n", members[0], hostData1.Cpha.Status)
		} else {
			fmt.Printf("host:%s:cpha:null\n", members[0])
		}

		hostData2, ok2 := checkStandalone(hosts, members[1], user, passw, su_passw, 22, verbose)

		clusterData.Hosts[members[1]] = hostData2

		if ok2 {
			fmt.Printf("host:%s:cpha:\"%s\"\n", members[1], hostData2.Cpha.Status)
		} else {
			fmt.Printf("host:%s:cpha:null\n", members[1])
		}

		if ok1 == false || ok2 == false {
			fmt.Printf("cluster:%s:routes_match:false\n", clustername)
			fmt.Printf("cluster:%s:ok:false\n", clustername)
			
			ok = false
		} else {
			_, host1OnlyRoutes, host2OnlyRoutes := CompareNetworks(hostData1.Routes, hostData2.Routes, verbose)
			
			ignoredRoutes := hosts.GetClusterIgnoredRoutes(clustername)
			host1Mismatch := false
			host2Mismatch := false
			
			var h1r sshtool.Routes
			var h2r sshtool.Routes
			
			if len(host1OnlyRoutes) != 0 {
				for _, r := range host1OnlyRoutes {
					if _, ok := ignoredRoutes[r.Net]; !ok {
						host1Mismatch = true
						h1r = append(h1r, r)
						//fmt.Printf("%-20s -> %-16s dev %s\n", r.Net, r.Gateway, r.Dev)
						//break
					}
				}
			}
			
			clusterData.Routes[members[0]] = h1r
			
			if len(host2OnlyRoutes) != 0 {
				for _, r := range host2OnlyRoutes {
					if _, ok := ignoredRoutes[r.Net]; !ok {
						host2Mismatch = true
						h2r = append(h2r, r)
						//break
					}
				}
			}

			clusterData.Routes[members[1]] = h2r
	
			if host1Mismatch || host2Mismatch {
				fmt.Printf("cluster:%s:routes_match:false\n", clustername)
				clusterData.Errors |= errRouteMismatch
			} else {
				fmt.Printf("cluster:%s:routes_match:true\n", clustername)
			}
			
			if (strings.Contains(hostData1.Cpha.Status, "active") || strings.Contains(hostData1.Cpha.Status, "standby")) && (strings.Contains(hostData2.Cpha.Status, "active") || strings.Contains(hostData2.Cpha.Status, "standby")) {
				if host1Mismatch || host2Mismatch {
					fmt.Printf("cluster:%s:ok:false\n", clustername)
					ok = false
				} else {
					fmt.Printf("cluster:%s:ok:true\n", clustername)
				}
			} else {
				fmt.Printf("cluster:%s:ok:false\n", clustername)
				clusterData.Errors |= errCphaStat
				ok = false
			}
		}
	} else {
		fmt.Printf("ERROR: cluster does not contain exactly two members\n")		
		ok = false
	}

	return clusterData, ok
}
