package cpu

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

// ignore stepping
var m = map[uint32]string{
	0x406e0: "Skylake",
	0x506e0: "Skylake",
	0x50650: "Skylake",
	0x806e0: "Kaby Lake",
	0x906e0: "Kaby Lake",
}

type microArch struct {
	Family int `toml:"famliy"`
	Model  int `toml:"model"`
}

var cpumapOnce sync.Once

// Concurrency-safe.
func NewCPUMap() map[uint32]string {
	cpumapOnce.Do(func() {
		var rt_viper = viper.New()
		var maps = map[string][]microArch{}

		// supported extensions are "json", "toml", "yaml", "yml", "properties", "props", "prop"
		rt_viper.SetConfigType("toml")
		rt_viper.SetConfigName("cpu_map")         // no need to include file extension
		rt_viper.AddConfigPath("/etc/rdtagent/")  // path to look for the config file in
		rt_viper.AddConfigPath("$HOME/rdtagent")  // call multiple times to add many search paths
		rt_viper.AddConfigPath("./etc/rdtagent/") // set the path of your config file
		err := rt_viper.ReadInConfig()

		if err != nil {
			fmt.Println(err)
		}

		rt_viper.Unmarshal(&maps)
		for k, mv := range maps {
			for _, v := range mv {
				sig := (v.Family>>4)<<20 + (v.Family&0xf)<<8 + (v.Model>>4)<<16 + (v.Model&0xf)<<4
				m[uint32(sig)] = k
			}
		}

	})
	return m
}

func GetMicroArch(sig uint32) string {
	s := sig & 0xFFFF0FF0
	NewCPUMap()
	if v, ok := m[s]; ok {
		return v
	}
	return ""
}
