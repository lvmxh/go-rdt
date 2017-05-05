package policy

import (
	"fmt"
	"testing"
)

func TestGetPolicy(t *testing.T) {
	t.Log("Testing get policy for broadwell cpu ... ")

	SetPolicyFilePath("../../../etc/policy.yaml")

	policy, err := GetPolicy("broadwell")

	if err != nil {
		t.Errorf("Failed to get policy for broadwell")
	}

	fmt.Println(policy.Gold.Size)
	if policy.Gold.Size != 14080 {
		t.Errorf("wrong policy for broadwell")
	}
}

func TestUpdatePolicy(t *testing.T) {
	t.Log("Testing update policy for broadwell cpu ... ")
	SetPolicyFilePath("../../../etc/policy.yaml")

	policy, err := GetPolicy("broadwell")

	if err != nil {
		t.Errorf("Failed to get policy for broadwell")
	}

	policy.Gold.Size = 1

	err = UpdatePolicy("broadwell", &policy)

	if err != nil {
		t.Errorf("Failed to Update policy for broadwell")
	}
	new_policy, err := GetPolicy("broadwell")

	if err != nil {
		t.Errorf("Failed to get policy for broadwell")
	}
	if new_policy.Gold.Size != 1 {
		t.Errorf("Failed to update policy")
	}

	// Write the origin value back
	new_policy.Gold.Size = 15
	UpdatePolicy("broadwell", &new_policy)
}
