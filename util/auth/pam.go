package auth

import (
	"bytes"
	"github.com/emicklei/go-restful"
	"net/http"
	"openstackcore-rdtagent/lib/proxyclient"
)

func PAMAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {

	u, p, ok := req.Request.BasicAuth()

	if !ok {
		resp.WriteErrorString(http.StatusBadRequest, "Malformed credentials\n")
		return
	}

	err := proxyclient.PAMAuthenticate(u, p)

	if err != nil {
		resp.AddHeader("WWW-Authenticate", "Basic realm=RMD")
		var buffer bytes.Buffer
		buffer.WriteString(err.Error())
		buffer.WriteString("\n")
		resp.WriteErrorString(http.StatusUnauthorized, buffer.String())
		return
	}

	chain.ProcessFilter(req, resp)
}
