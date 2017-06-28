package main

// // No C code needed.
import "C"

import (
	ipolicy "./internal/policy"
	"./internal/workload"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
)

type algorithmSample struct {
}

// Get the policy a file. this is just a fake read from file.
type SLA struct {
	size, percentage int
	pin_core         bool
}

func policy_get_from_file(name string) map[string]SLA {
	var p = map[string]SLA{
		name: {14080, 20, false},
	}
	return p
}

func (as algorithmSample) Enforce(wl workload.RDTWorkLoad, policy string) {
	/* Use simplejson to parser the policy json string */
	js, _ := simplejson.NewJson([]byte(policy))
	name, _ := js.Get("policy").Get("Name").String()
	if name == "dilangov_dynamic" {
		variance, _ := js.Get("policy").Get("Variance").Int()
		pu, _ := js.Get("policy").Get("PeakUsage").Int()
		ov, _ := js.Get("policy").Get("Overcommit").Bool()
		fmt.Println("Example for Enforce policy:", name, variance, pu, ov)
		return
	}

	/* Use encoding/json to parser the policy json string to interface*/
	var f interface{}
	json.Unmarshal([]byte(policy), &f)
	m := f.(map[string]interface{})
	p := m["policy"].(map[string]interface{})

	if p["name"] == "v1_strict" {
		sla := policy_get_from_file("gold")
		fmt.Println("Example for Enforce policy:", p["name"], sla)
		return
	}

	/* Use encoding/json to parser the policy json string to struct*/
	if p["name"] == "3_party" {
		var p3 ipolicy.Policy3P
		json.Unmarshal([]byte(policy), &p3)
		fmt.Println("Example for Enforce policy:",
			p3.Policy.Name, p3.Policy.Data)
		return
	}
}

var Algorithm algorithmSample

func main() {
	fmt.Println("This is an algorithm example.")
}
