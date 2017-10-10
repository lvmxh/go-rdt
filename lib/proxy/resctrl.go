package proxy

import (
	"openstackcore-rdtagent/lib/resctrl"
)

type ProxyRequest struct {
	Name string
	Res  resctrl.ResAssociation
}

func (_ *Proxy) Commit(r ProxyRequest, dummy *int) error {
	return resctrl.Commit(&r.Res, r.Name)
}

func (_ *Proxy) DestroyResAssociation(grpName string, dummy *int) error {
	return resctrl.DestroyResAssociation(grpName)
}

func (_ *Proxy) RemoveTasks(tasks []string, dummy *int) error {
	return resctrl.RemoveTasks(tasks)
}

func (_ *Proxy) EnableCat(dummy *int, result *bool) error {
	*result = resctrl.EnableCat()
	return nil
}
