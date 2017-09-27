package auth

import (
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/msteinert/pam"
	"net/http"
)

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

func PamAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	u, p, ok := req.Request.BasicAuth()
	if !ok {
		resp.WriteErrorString(http.StatusBadRequest, "Malformed credentials")
		return
	}

	c := Credentials{
		Username: u,
		Password: p,
	}

	tx, err := pam.StartFunc("rmd", c.Username, c.PAMResponseHandler)
	if err != nil {
		resp.WriteErrorString(http.StatusInternalServerError, "Failed to load authentication module")
		return
	}

	err = tx.Authenticate(0)
	if err != nil {
		resp.AddHeader("WWW-Authenticate", "Basic realm=RMD")
		resp.WriteErrorString(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	chain.ProcessFilter(req, resp)
}
