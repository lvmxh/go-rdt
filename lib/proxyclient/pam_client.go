package proxyclient

import (
	"openstackcore-rdtagent/lib/proxy"
)

func PAMAuthenticate(user string, pass string) error {

	req := proxy.PAMRequest{user, pass}
	err := proxy.Client.Call("Proxy.PAMAuthenticate", req, nil)
	return err
}
