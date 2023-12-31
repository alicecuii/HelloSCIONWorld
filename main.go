// Copyright 2018 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alicecuii/HelloSCIONWorld/regionrule"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"inet.af/netaddr"
)

func usageErr(msg string) {
	if msg != "" {
		fmt.Println("\nError:", msg)
	}
	os.Exit(2)
}

func checkUsageErr(err error) {
	if err != nil {
		usageErr(err.Error())
	}
}

func main() {
	var err error
	// get local and remote addresses from program arguments:
	var (
		listen        pan.IPPortValue
		rule          string
		interactive   bool
		sequence      string
		preference    string
		rules         []regionrule.Rule
		remoteAddr    string
		permitted_ISD []int
	)
	rule_names := []string{}
	rules, err = regionrule.GetRules()
	// Iterate through the list of rules and collect the Preference values
	for _, rule1 := range rules {
		rule_names = append(rule_names, rule1.Name)
	}
	// Create a map where the keys are rule names and the values are preferences
	rulePreferences := make(map[string]string)
	ruleISDs := make(map[string][]int)
	// Populate the map
	for _, rule1 := range rules {
		rulePreferences[rule1.Name] = rule1.Preference
		// Check if the rule name already exists in the map
		if _, ok := ruleISDs[rule1.Name]; !ok {
			// If it doesn't exist, create a new entry with an empty slice
			ruleISDs[rule1.Name] = []int{}
		}
		// Append the permitted ISD values to the slice
		ruleISDs[rule1.Name] = append(ruleISDs[rule1.Name], rule1.Permitted_ISD...)
	}

	flag.Var(&listen, "listen", "[Server] local IP:port to listen on")
	flag.StringVar(&remoteAddr, "remote", "", "[Client] Remote (i.e. the server's) SCION Address (e.g. 17-ffaa:1:1,[127.0.0.1]:12345)")
	flag.StringVar(&rule, "rule", "", "Preference sorting order for paths. "+
		"Comma-separated list of available sorting options: "+
		strings.Join(rule_names, "|"))

	flag.BoolVar(&interactive, "i", false, "Interactive path selection, prompt to choose path")
	flag.StringVar(&sequence, "sequence", "", "Sequence of space separated hop predicates to specify path")
	flag.StringVar(&preference, "preference", "", "Preference sorting order for paths. "+
		"Comma-separated list of available sorting options: "+
		strings.Join(pan.AvailablePreferencePolicies, "|"))

	count := flag.Uint("count", 1, "[Client] Number of messages to send")
	flag.Parse()

	if preference != "" && rule != "" {
		check(fmt.Errorf("either specify -preference or -rule"))
	} else if preference == "" && rule != "" {
		preference = rulePreferences[rule]
		permitted_ISD = ruleISDs[rule]
	}
	fmt.Println("preference: ", preference)
	policy, err := pan.PolicyFromCommandline(sequence, preference, interactive)
	checkUsageErr(err)

	if (listen.Get().Port() > 0) == (len(remoteAddr) > 0) {
		check(fmt.Errorf("either specify -listen for server or -remote for client"))
	}

	if listen.Get().Port() > 0 {
		err = runServer(listen.Get())
		check(err)
	} else {
		err = runClient(remoteAddr, int(*count), policy, permitted_ISD)
		check(err)
	}
}

func runServer(listen netaddr.IPPort) error {
	conn, err := pan.ListenUDP(context.Background(), listen, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Println(conn.LocalAddr())

	buffer := make([]byte, 16*1024)
	for {
		n, from, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		data := buffer[:n]
		fmt.Printf("Received %s: %s\n", from, data)
		msg := fmt.Sprintf("take it back! %s", time.Now().Format("15:04:05.0"))
		n, err = conn.WriteTo([]byte(msg), from)
		if err != nil {
			return err
		}
		fmt.Printf("Wrote %d bytes.\n", n)
	}
}

func runClient(address string, count int, policy pan.Policy, isds []int) error {
	addr, err := pan.ResolveUDPAddr(address)
	if err != nil {
		fmt.Println("server address error")
		return err
	}
	//Select path to control connection
	pathSelector := pan.NewDefaultSelector()
	conn, err := pan.DialUDP(context.Background(), netaddr.IPPort{}, addr, policy, pathSelector)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(isds)
	fmt.Print("Chosen path: ")
	fmt.Println(pathSelector.Path())
	defer conn.Close()

	for i := 0; i < count; i++ {
		nBytes, err := conn.Write([]byte(fmt.Sprintf("hello world %s", time.Now().Format("15:04:05.0"))))
		if err != nil {
			return err
		}
		fmt.Printf("Wrote %d bytes.\n", nBytes)

		buffer := make([]byte, 16*1024)
		if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			return err
		}
		n, err := conn.Read(buffer)
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue
		} else if err != nil {
			return err
		}
		data := buffer[:n]
		fmt.Printf("Received reply: %s\n", data)
	}
	return nil
}

// Check just ensures the error is nil, or complains and quits
func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, "Fatal error:", e)
		os.Exit(1)
	}
}
