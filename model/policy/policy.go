package policy

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

// FIXME move this to configure file
var yaml_path = "/etc/rdtagent/policy.yaml"

type Attr map[string]string

type Policy map[string][]Attr

type CATConfig struct {
	Catpolicy map[string][]Policy `yaml:"catpolicy"`
}

var config *CATConfig
var lock sync.Mutex

// For testing
func SetPolicyFilePath(path string) {
	yaml_path = path
}

func LoadPolicy() (*CATConfig, error) {
	r, err := ioutil.ReadFile(yaml_path)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	c := CATConfig{}
	err = yaml.Unmarshal(r, &c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &c, err
}

func GetPlatformPolicy(cpu string) ([]Policy, error) {

	lock.Lock()
	defer lock.Unlock()
	var err error

	if config == nil {
		config, err = LoadPolicy()
		if err != nil {
			return []Policy{}, err
		}
	}

	p, ok := config.Catpolicy[cpu]

	if !ok {
		return []Policy{}, fmt.Errorf("Error while get platform policy: %s", cpu)
	}

	return p, nil
}

// return a map of the policy has
func GetPolicy(cpu, policy string) (map[string]string, error) {
	m := make(map[string]string)

	platform, err := GetPlatformPolicy(cpu)

	if err != nil {
		return m, fmt.Errorf("Can not find specified platform policy.")
	}

	var policyCandidate []Policy

	for _, p := range platform {
		_, ok := p[policy]
		if ok {
			policyCandidate = append(policyCandidate, p)
		}
	}
	if len(policyCandidate) == 1 {
		for _, item := range policyCandidate[0][policy] {
			// merge to one map
			for k, v := range item {
				m[k] = v
			}
		}
		return m, nil
	} else {
		return m, fmt.Errorf("Can not find specified policy %s", policy)
	}
}
