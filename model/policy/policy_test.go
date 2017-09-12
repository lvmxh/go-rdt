package policy

import (
	"testing"
)

func TestGetPolicy(t *testing.T) {
	t.Log("Testing get policy for broadwell cpu ... ")

	SetPolicyFilePath("../../etc/rdtagent/policy.yaml")

	_, err := GetPolicy("broadwell", "gold")

	if err != nil {
		t.Errorf("Failed to get gold policy for broadwell")
	}

	_, err = GetPolicy("broadwell", "foo")

	if err == nil {
		t.Errorf("Error should be return as no foo policy")
	}
}
