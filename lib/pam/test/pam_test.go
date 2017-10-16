package test

import (
	"openstackcore-rdtagent/lib/pam"
	"testing"
)

func TestPAMStartFunc(t *testing.T) {
	_, err := pam.PAMStartFunc("", "", nil)
	if err != nil {
		t.Error(err)
	}
}

func TestPAMTxAuthenticate(t *testing.T) {
	// valid credential
	c := pam.Credential{"user", "user1"}

	// valid service name
	service := "rmd"

	tx, err := pam.PAMStartFunc(service, c.Username, c.PAMResponseHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = pam.PAMTxAuthenticate(tx)
	if err != nil {
		t.Error(err)
	}
}

func TestPAMAuthenticate(t *testing.T) {

	// Litmus test start func
	_, err := pam.PAMStartFunc("", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		username      string
		password      string
		description   string
		desiredResult string
	}{
		{"user", "user1", "Valid Berkeley db user", ""},
		{"x", "y", "Invalid Berkeley db user", "User not known to the underlying authentication module"},
		{"user", "user", "Incorrect Berkeley db user", "Authentication failure"},
		{"common", "common", "Valid linux user", ""},
		{"a", "b", "Invalid linux user", "User not known to the underlying authentication module"},
		{"common", "c", "Incorrect linux user", "User not known to the underlying authentication module"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c := pam.Credential{testCase.username, testCase.password}
			err := c.PAMAuthenticate()
			if testCase.desiredResult == "" {
				if err != nil {
					t.Error(err)
				}
			} else {
				if err == nil {
					t.Error("No error detected as desired. Please check test inputs")
				}
				if err.Error() != testCase.desiredResult {
					t.Error(err)
				}
			}
		})
	}
}
