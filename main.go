package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type HealthCheck struct {
	Node        string   `json:"Node"`
	CheckID     string   `json:"CheckID"`
	Name        string   `json:"Name"`
	Status      string   `json:"Status"`
	Notes       string   `json:"Notes"`
	Output      string   `json:"Output"`
	ServiceID   string   `json:"ServiceID"`
	ServiceName string   `json:"ServiceName"`
	ServiceTags []string `json:"ServiceTags"`
	CreateIndex int      `json:"CreateIndex"`
	ModifyIndex int      `json:"ModifyIndex"`
}

var version string
var githash string
var buildstamp string

var Critical []HealthCheck
var Warning []HealthCheck
var Passing []HealthCheck
var Checks []HealthCheck

func displayVersion() {
	versionText := `Icinga2/Nagios Consul Plugin

Version: %s
    Git: %s
  Build: %s
`

	fmt.Println(fmt.Sprintf(versionText, version, githash, buildstamp))
}

func main() {
	hostname, err := os.Hostname()

	versionFlag := flag.Bool("version", false, "display version information")
	hostFlag := flag.String("host", "http://localhost:8500", "consul http(s) address with protocol")
	nodeFlag := flag.String("node", hostname, "the current hostname")
	serviceFlag := flag.String("service", "consul", "consul service id")

	flag.Parse()

	if *versionFlag == true {
		displayVersion()
		os.Exit(0)
	}

	host := *hostFlag
	node := *nodeFlag
	service := *serviceFlag

	if host == "" {
		fmt.Println("--host must be set")
		os.Exit(3)
	}

	if service == "" {
		fmt.Println("--service must be set")
		os.Exit(3)
	}

	consul_url := fmt.Sprintf("%s/v1/health/node/%s", host, node)

	resp, err := http.Get(consul_url)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	if err := json.Unmarshal(body, &Checks); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	for _, check := range Checks {
		if check.Status == "critical" && check.ServiceID == service {
			Critical = append(Critical, check)
		} else if check.Status == "warning" && check.ServiceID == service {
			Warning = append(Warning, check)
		} else if check.Status == "passing" && check.ServiceID == service {
			Passing = append(Passing, check)
		}
	}

	if len(Critical) == 0 && len(Warning) == 0 && len(Passing) == 0 {
		fmt.Println(fmt.Sprintf("WARNING - No checks for service '%s' found!", service))
		os.Exit(1)
	}

	var code = 3
	var text = "UNKNOWN"
	var messages []string
	var outputs []string
	var notes []string

	if len(Critical) > 0 {
		code = 2
		for _, check := range Critical {
			text = "CRITICAL - "
			messages = append(messages, check.Name)
			if check.Output != "" {
				outputs = append(outputs, check.Output)
			}
			if check.Notes != "" {
				notes = append(notes, check.Notes)
			}
		}
	} else if len(Warning) > 0 {
		code = 1
		for _, check := range Warning {
			text = "WARNING - "
			messages = append(messages, check.Name)
			if check.Output != "" {
				outputs = append(outputs, check.Output)
			}
			if check.Notes != "" {
				notes = append(notes, check.Notes)
			}
		}
	} else if len(Passing) > 0 {
		code = 0
		for _, check := range Passing {
			text = "OK - "
			messages = append(messages, check.Name)
			if check.Output != "" {
				outputs = append(outputs, check.Output)
			}
			if check.Notes != "" {
				notes = append(notes, check.Notes)
			}
		}
	} else {
		code = 3
		text = "UNKNOWN - No Critical, Warning or Passing Checks Found - This should not happen."
	}

	text += "Checks: " + strings.Join(messages, ", ")

	if len(outputs) > 0 {
		text += "\nOutputs: "
		for _, out := range outputs {
			text += "\n  " + out
		}
	}

	if len(notes) > 0 {
		text += "\nNotes:"
		for _, out := range notes {
			text += "\n  " + out
		}
	}

	fmt.Println(text)
	os.Exit(code)
}
