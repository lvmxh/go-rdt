package pam

import (
	"errors"
	"fmt"
	"github.com/intel/rmd/lib/pam/config"
	"github.com/msteinert/pam"
)

// Credential represents user provided credential
type Credential struct {
	Username string
	Password string
}

// PAMResponseHandler handles communication between PAM client and PAM module
func (c Credential) PAMResponseHandler(s pam.Style, msg string) (string, error) {
	switch s {
	case pam.PromptEchoOff:
		return c.Password, nil
	case pam.PromptEchoOn:
		fmt.Println(msg)
		return c.Password, nil
	case pam.ErrorMsg:
		fmt.Println(msg)
		return "", nil
	case pam.TextInfo:
		fmt.Println(msg)
		return "", nil
	}
	return "", errors.New("Unrecognized message style")
}

// TxAuthenticate does transaction
func TxAuthenticate(transaction *pam.Transaction) error {
	err := transaction.Authenticate(0)
	return err
}

// PAMAuthenticate does PAM authentication
func (c Credential) PAMAuthenticate() error {
	tx, err := c.PAMStartFunc()
	if err != nil {
		return err
	}
	err = TxAuthenticate(tx)
	return err
}

// StartFunc does start
func StartFunc(service string, user string, handler func(pam.Style, string) (string, error)) (*pam.Transaction, error) {
	tx, err := pam.StartFunc(service, user, handler)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// PAMStartFunc establishes connection to PAM module
func (c Credential) PAMStartFunc() (*pam.Transaction, error) {
	return StartFunc(config.GetPAMConfig().Service, c.Username, c.PAMResponseHandler)
}
