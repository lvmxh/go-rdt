package policy

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"sync"

	appConf "openstackcore-rdtagent/app/config"
	"openstackcore-rdtagent/lib/cpu"
)

type Attr map[string]string

type Policy map[string][]Attr

type CATConfig struct {
	Catpolicy map[string][]Policy `yaml:"catpolicy"`
}

var config *CATConfig
var lock sync.Mutex

func LoadPolicy() (*CATConfig, error) {
	appconf := appConf.NewConfig()
	r, err := ioutil.ReadFile(appconf.Def.PolicyPath)
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

// GetdefaultPlatformPolicy, wrapper for GetPlatformPolicy
func GetDefaultPlatformPolicy() ([]Policy, error) {
	cpu := cpu.GetMicroArch(cpu.GetSignature())
	if cpu == "" {
		return []Policy{}, fmt.Errorf("Unknow platform, please update the cpu_map.toml conf file")
	}

	return GetPlatformPolicy(strings.ToLower(cpu))
}

// GetPolicy return a map of the policy of the host
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

// GetDefaultPolicy return a map of the default policy of the host
func GetDefaultPolicy(policy string) (map[string]string, error) {
	cpu := cpu.GetMicroArch(cpu.GetSignature())
	if cpu == "" {
		return map[string]string{}, fmt.Errorf("Unknow platform, please update the cpu_map.toml conf file")
	}
	return GetPolicy(cpu, policy)
}
