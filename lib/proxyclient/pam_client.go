package proxyclient

import (
	"openstackcore-rdtagent/lib/proxy"
)

// PAMAuthenticate leverage PAM to do authentication
func PAMAuthenticate(user string, pass string) error {

	req := proxy.PAMRequest{user, pass}
	return proxy.Client.Call("Proxy.PAMAuthenticate", req, nil)
}
