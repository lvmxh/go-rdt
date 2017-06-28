package main

import (
	"./internal/policy"
	"./internal/workload"
	"encoding/json"
	"fmt"
	"plugin"
)

type algorithm interface {
	Enforce(wl workload.RDTWorkLoad, policy string)
}

// emulate a server get a policy json context from user request
// The routine of a RDT perform as follow:
// http rest api -> load the plugin -> plugin enfore.
func policy_api(name string) {
	// var workload interface{}
	pol := ""
	if name == "v1_strict" {
		pol = `{"policy": {"name":"v1_strict"}}`
	}
	if name == "3_party" {
		pol = `{"policy": {"name":"3_party", "data": "abc"}}`
	}
	// dilangov's dynamic policy is from user request.
	if name == "dilangov_dynamic" {
		var policy policy.Policy = policy.Policy{
			"dilangov_dynamic", 10, 20, true}
		p, _ := json.Marshal(policy)
		pol = fmt.Sprintf("{\"policy\": %s}", string(p))
	}
	// In real case, the "v1_strict" and "dilangov_dynamic" plugin should
	// be separated. We should define a v1_strict.so plugin for v1_strict.
	// We should also define a dilangov_dynamic.so plugin for dilangov_dynamic.
	// Here both of them call sample.so
	pl, err := plugin.Open("sample.so")
	if err != nil {
		fmt.Println(err)
	}
	algSymbol, err := pl.Lookup("Algorithm")
	if err != nil {
		fmt.Println(err)
	}
	alg, ok := algSymbol.(algorithm)
	if !ok {
		fmt.Println("unexpected type from module symbol")
	}
	wl := workload.RDTWorkLoad{}
	alg.Enforce(wl, pol)
}

func ExampleAlgorithm() {
	fmt.Println("Example for Enforce")
	policy_api("dilangov_dynamic")
	policy_api("v1_strict")
	policy_api("3_party")
	// Output:
	// Example for Enforce
	// Example for Enforce policy: dilangov_dynamic 10 20 true
	// Example for Enforce policy: v1_strict map[gold:{14080 20 false}]
	// Example for Enforce policy: 3_party abc
}
