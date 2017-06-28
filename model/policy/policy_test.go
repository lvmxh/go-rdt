package policy

import (
	"testing"
)

func TestGetPolicy(t *testing.T) {
	t.Log("Testing get policy for broadwell cpu ... ")

	SetPolicyFilePath("../../../etc/policy.yaml")

	policy, err := GetPolicy("broadwell", "gold")

	if err != nil {
		t.Errorf("Failed to get gold policy for broadwell")
	}

	if policy["peakusage"] != "14080" {
		t.Errorf("Error peakusage in gold policy for broadwell")
	}

	policy, err = GetPolicy("broadwell", "foo")

	if err == nil {
		t.Errorf("Error should be return as no foo policy")
	}
}
