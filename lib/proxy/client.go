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

// GetResAssociation returns all resource group association
func GetResAssociation() map[string]*resctrl.ResAssociation {
	return resctrl.GetResAssociation()
}

// GetRdtCosInfo returns RDT infromations
func GetRdtCosInfo() map[string]*resctrl.RdtCosInfo {
	return resctrl.GetRdtCosInfo()
}

// IsIntelRdtMounted will check if resctrl mounted or not
func IsIntelRdtMounted() bool {
	return resctrl.IsIntelRdtMounted()
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

// RemoveTasks moves tasks to default resource group
func RemoveTasks(tasks []string) error {
	// TODO how to get error reason
	return Client.Call("Proxy.RemoveTasks", tasks, nil)
}

// EnableCat() enable cat feature on host
func EnableCat() error {
	var result bool
	if err := Client.Call("Proxy.EnableCat", 0, &result); err != nil {
		return err
	}
	if result {
		return nil
	} else {
		return fmt.Errorf("Can not enable cat")
	}
}
