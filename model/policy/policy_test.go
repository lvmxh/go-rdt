package policy

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestGetPolicy(t *testing.T) {
	t.Log("Testing get policy for broadwell cpu ... ")

	pflag.String("address", "", "Listen address")
	pflag.Int("tlsport", 0, "TLS listen port")
	pflag.BoolP("debug", "d", false, "Enable debug")
	pflag.String("unixsock", "", "Unix sock file path")
	pflag.Int("debugport", 0, "Debug listen port")
	pflag.String("conf-dir", "", "Directy of config file")

	pflag.Parse()

	viper.Set("default.policypath", "../../etc/rdtagent/policy.yaml")
	_, err := GetPolicy("broadwell", "gold")

	if err != nil {
		t.Errorf("Failed to get gold policy for broadwell")
	}

	_, err = GetPolicy("broadwell", "foo")

	if err == nil {
		t.Errorf("Error should be return as no foo policy")
	}
}
