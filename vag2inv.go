// Generate Ansible inventory file from Vagrant.
//
//
// Copyright 2015 William Yeh <william.pjyeh@gmail.com>. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"

	"regexp"
	"strings"

	"bufio"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/docopt/docopt-go"
)

const VERSION_NUMBER string = "0.1"

type host_info struct {
	name     string
	addr     string
	port     string
	user     string
	key_file string
}

/* `vagrant status`

^([^\s]+)\s+    ([^(]+)  \s+\(.+$
default                   not created (virtualbox)
node1                     running (virtualbox)
node2                     running (virtualbox)

*/
var REGEX_VAGRANT_STATUS = regexp.MustCompile(`^([^\s]+)\s+([^(]+)\s+\(.+$`)

var REGEX_HOST_ENTRY_HEAD = regexp.MustCompile(`^(Host)\s+(.+)$`)
var REGEX_HOST_ENTRY_DETAILS = regexp.MustCompile(`^\s+([^\s]+)\s+(.+)$`)

var REGEX_PRIVATE_KEY_PATH = regexp.MustCompile(`^(.+)/(.vagrant/machines/.+/private_key)$`)

var REGEX_IFCONFIG_ETH0_IPV4 = regexp.MustCompile(`^\s+inet addr:([^\s]+)\s.+$`)

const USAGE string = `Generate Ansible inventory file from Vagrant.

Usage:
  vag2inv  [options]  <inventory_filename>
  vag2inv  --help
  vag2inv  --version

Options:
  -d, --stdout                Also dump to stdout.
  -f, --force                 Force overwrite inventory file;
                                [default: false].
  --vm                        Compatible for Ansible control machine that resides in VM;
                                [default: false].
  -p <dir>, --prefix <dir>    Rewrite the prefix part of the private key's path.

`

func main() {
	arguments := process_cmdline()
	//fmt.Println(arguments)

	num_running, num_non_running := collect_vagrant_status_output()
	//fmt.Printf("Running: %d  Non-running: %d\n", num_running, num_non_running)
	if num_running < 1 || num_non_running > 0 {
		fmt.Println("Abort: not all boxes are running.")
		os.Exit(2)
	}

	ssh_config := collect_ssh_config_output()
	//fmt.Println(ssh_config)

	//fix_for_vm_node(&ssh_config)
	//fmt.Println(ssh_config)

	output_file(arguments, ssh_config)
}

// It parses and validates cmdline args
func process_cmdline() map[string]interface{} {
	arguments, _ := docopt.Parse(USAGE, nil, true, VERSION_NUMBER, false)

	// validate inventory file
	filename := arguments["<inventory_filename>"].(string)

	if _, err := os.Stat(filename); err == nil {
		if !arguments["--force"].(bool) {
			fmt.Printf("Error: output file already exists: %s", filename)
			os.Exit(1)
		}
		os.Remove(filename)
	}

	return arguments
}

// It collects and parses the "vagrant status".
func collect_vagrant_status_output() (int, int) {
	out, err := exec.Command("vagrant", "status").Output()
	if err != nil {
		log.Fatal("Error inspecting Vagrant status; ", err)
		//os.Exit(1)
	}

	var num_running = 0
	var num_non_running = 0
	for _, line := range strings.Split(string(out), "\n") { // for each line
		if result := REGEX_VAGRANT_STATUS.FindStringSubmatch(line); result != nil {
			//fmt.Println("---> ", line)
			//fmt.Printf("1: %s  2: %s\n", result[1], result[2])
			if result[2] == "running" {
				num_running += 1
			} else {
				num_non_running += 1
			}
		}
	}

	//fmt.Printf("Running: %d  Non-running: %d\n", num_running, num_non_running)
	return num_running, num_non_running
}

// It collects and parses the "vagrant ssh-config";
// aborts if not all boxes are in running state.
func collect_ssh_config_output() []host_info {
	out, err := exec.Command("vagrant", "ssh-config").Output()
	if err != nil {
		log.Fatal("Error inspecting Vagrant ssh-config; ", err)
		//os.Exit(1)
	}
	// delete all "\r" characters, if any (Windows)
	new_out := strings.Replace(string(out), "\r", "", -1)
	//fmt.Printf("in all caps: %q\n", new_out)

	var hosts []host_info
	for _, host_entry := range strings.Split(new_out, "\n\n") { // for each host entry
		host_entry_trim := strings.TrimSpace(host_entry)
		if len(host_entry_trim) < 1 { // skip if empty (usually the last line)
			continue
		}

		//fmt.Println("===== HOST ====== ", len(host_entry_trim))
		//fmt.Println(host_entry_trim)
		info := parse_host_entry(host_entry_trim)
		//fmt.Println(info)
		hosts = append(hosts, info)
	}

	return hosts
}

// It parses a single entry of the "vagrant ssh-config".
func parse_host_entry(host_entry string) host_info {
	info := host_info{}

	for _, line := range strings.Split(host_entry, "\n") { // for each line
		if result := REGEX_HOST_ENTRY_HEAD.FindStringSubmatch(line); result != nil {
			info.name = result[2]
		} else if result := REGEX_HOST_ENTRY_DETAILS.FindStringSubmatch(line); result != nil {
			key := result[1]
			value := result[2]

			switch key {
			case "HostName":
				info.addr = value
			case "User":
				info.user = value
			case "Port":
				info.port = value
			case "IdentityFile":
				info.key_file = value
			}
		}
	}

	//fmt.Println(info)
	return info
}

// It queries internal IPv4 address of specific Vagrant box
// via "ifconfig" on "enp0s3" or "eth0".
//
// NOTE: Linux with systemd may use the "enp0s3" name:
//   - http://askubuntu.com/questions/704035/no-eth0-listed-in-ifconfig-a-only-enp0s3-and-lo
//   - http://linuxconfig.org/configuring-network-interface-with-static-ip-address-on-rhel-7
//
func query_host_eth0_ipv4(host_name string) string {

	var ifconfig_output []byte
	for _, ifname := range []string{"enp0s3", "eth0"} {
		out, err := exec.Command("vagrant", "ssh", host_name, "-c", "ifconfig "+ifname).Output()
		if err == nil {
			ifconfig_output = out
			break
		}
	}

	if ifconfig_output == nil {
		//fmt.Println("ERR: ", out)
		log.Fatal("Error inspecting 'vagrant ssh ", host_name, " -c ' ; ")
		//os.Exit(1)
	}

	for _, line := range strings.Split(string(ifconfig_output), "\n") { // for each line
		if result := REGEX_IFCONFIG_ETH0_IPV4.FindStringSubmatch(line); result != nil {
			return result[1]
		}
	}
	return "127.0.0.1"
}

// It outputs inventory contents to external file, and optionally to stdout.
func output_file(arguments map[string]interface{}, ssh_config []host_info) {
	to_stdout := arguments["--stdout"].(bool)
	inv_file := arguments["<inventory_filename>"].(string)

	// open output file
	file, err := os.Create(inv_file)
	if err != nil {
		log.Fatal("Error writing inventory file; ", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)

	output_header(writer, to_stdout)

	for _, host_entry := range ssh_config {
		//fmt.Println(host_entry)

		var str string

		if arguments["--vm"].(bool) {
			internal_addr := query_host_eth0_ipv4(host_entry.name)
			//fmt.Println(internal_addr)
			if internal_addr != "127.0.0.1" {
				host_entry.addr = internal_addr
				host_entry.port = "22"
			}

			str = fmt.Sprintf(
				"%s ansible_ssh_host=%s ansible_host=%s ansible_ssh_port=%s ansible_port=%s ansible_ssh_user=%s ansible_user=%s ansible_ssh_pass=vagrant\n",
				host_entry.name,
				host_entry.addr, host_entry.addr,
				host_entry.port, host_entry.port,
				host_entry.user, host_entry.user)

		} else {
			private_key_file := host_entry.key_file
			if arguments["--prefix"] != nil {
				//var REGEX_PRIVATE_KEY_PATH = regexp.MustCompile(`^(.+)/(.vagrant/machines/.+/private_key)$`)
				if result := REGEX_PRIVATE_KEY_PATH.FindStringSubmatch(private_key_file); result != nil {
					private_key_file = path.Join(arguments["--prefix"].(string), result[2])
				}
			}

			// Ansible 2.0 has deprecated the “ssh” in ansible_ssh_* variables.
			// @see http://docs.ansible.com/ansible/intro_inventory.html
			str = fmt.Sprintf(
				"%s ansible_ssh_host=%s ansible_host=%s ansible_ssh_port=%s ansible_port=%s ansible_ssh_user=%s ansible_user=%s ansible_ssh_private_key_file=%s\n",
				host_entry.name,
				host_entry.addr, host_entry.addr,
				host_entry.port, host_entry.port,
				host_entry.user, host_entry.user,
				private_key_file)
		}

		// output...
		fmt.Fprintln(writer, str)
		if to_stdout {
			fmt.Println(str)
		}
	}
	writer.Flush()
}

// It outputs useful info at the beginning of inventory file.
func output_header(writer *bufio.Writer, to_stdout bool) {
	str := fmt.Sprintf(
		"# Generated by vag2inv\n# Cmdline: %s\n# @see https://github.com/William-Yeh/vag2inv\n\n",
		strings.Join(os.Args, " "))

	fmt.Fprint(writer, str)
	if to_stdout {
		fmt.Print(str)
	}
}
