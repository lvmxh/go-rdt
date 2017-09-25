package proxy

import (
	"fmt"
	"net/rpc"
	"openstackcore-rdtagent/lib/resctrl"
)

// Client is the connection to rpc server
var Client *rpc.Client

// ConnecRPCServer by a pipe pair
// Be care about this method usage, it can only be called once while
// we start RMD API server, sync.once could be one choice, developer
// should control it.
func ConnectRPCServer(in PipePair) error {
	Client = rpc.NewClient(&in)
	if Client == nil {
		return fmt.Errorf("Faild to connect rpc server")
	}
	return nil
}

// Commit resctrl.ResAssociation with given name
func Commit(r *resctrl.ResAssociation, name string) error {
	// TODO how to get error reason
	req := ProxyRequest{name, *r}
	// Add checking before using client and do reconnect
	return Client.Call("Proxy.Commit", req, nil)
}

// DestroyResAssociation by resource group name
func DestroyResAssociation(name string) error {
	// TODO how to get error reason
	// Add checking before using client and do reconnect
	return Client.Call("Proxy.DestroyResAssociation", name, nil)
}
