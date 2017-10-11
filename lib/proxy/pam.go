package proxy

import (
	"errors"
	"fmt"
	"github.com/msteinert/pam"
	"openstackcore-rdtagent/lib/proxy/config"
)

// PAMRequest is request from rpc client
type PAMRequest struct {
	User string
	Pass string
}

// Credential represents user provided credential
type Credential struct {
	Username string
	Password string
}

// PAMResponseHandler is handler for PAM Authentication of rpc server
func (c Credential) PAMResponseHandler(s pam.Style, msg string) (string, error) {
	switch s {
	case pam.PromptEchoOff:
		return c.Password, nil
	case pam.PromptEchoOn:
		fmt.Println(msg)
		return c.Password, nil
	case pam.ErrorMsg:
		fmt.Errorf(msg)
		return "", nil
	case pam.TextInfo:
		fmt.Println(msg)
		return "", nil
	}
	return "", errors.New("Unrecognized message style")
}

// PAMAuthenticate does PAM authenticate
func (*Proxy) PAMAuthenticate(request PAMRequest, dummy *int) error {

	c := Credential{
		Username: request.User,
		Password: request.Pass,
	}

	tx, err := pam.StartFunc(config.GetPAMConfig().Service, request.User, c.PAMResponseHandler)
	if err != nil {
		return err
	}

	err = tx.Authenticate(0)
	return err
}
