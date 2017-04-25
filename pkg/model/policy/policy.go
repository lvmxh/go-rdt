package policy

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

var yaml_path = "/etc/rdtagent/policy.yaml"

type PolicyRule struct {
	Dpdk  string
	Other string
}
type PolicyType struct {
	Size       uint
	Percentage uint
	Rule       PolicyRule
	Pin_core   bool
}

type Policy struct {
	Gold   PolicyType
	Silver PolicyType
	Copper PolicyType
}

type CATConfig struct {
	Catpolicy map[string]*Policy `yaml:"catpolicy"`
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

// return a copy of policy, not a pointer
func GetPolicy(cpu string) (Policy, error) {
	lock.Lock()
	defer lock.Unlock()
	var err error

	if config == nil {
		config, err = LoadPolicy()
		if err != nil {
			return Policy{}, err
		}
	}

	p, ok := config.Catpolicy[cpu]
	if ok != true {
		return Policy{}, errors.New("cpu doen't supported")
	}

	return *p, nil
}

// Write back policy to yaml file
func UpdatePolicy(cpu string, p *Policy) error {
	lock.Lock()
	defer lock.Unlock()
	if config == nil {
		return errors.New("empty config file")
	}

	_, ok := config.Catpolicy[cpu]
	if ok != true {
		return errors.New("cpu doen't supported")
	}

	config.Catpolicy[cpu] = p

	d, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("error: %v", err)
		return err
	}
	return ioutil.WriteFile(yaml_path, d, 0644)
}
