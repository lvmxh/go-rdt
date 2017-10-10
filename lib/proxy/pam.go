package proxy

import (
	"errors"
	"fmt"
	"github.com/msteinert/pam"
	"openstackcore-rdtagent/lib/proxy/config"
)

type PAMRequest struct {
	User string
	Pass string
}

type Credentials struct {
	Username string
	Password string
}

func (c Credentials) PAMResponseHandler(s pam.Style, msg string) (string, error) {
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

func (_ *Proxy) PAMAuthenticate(request PAMRequest, dummy *int) error {

	c := Credentials{
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
