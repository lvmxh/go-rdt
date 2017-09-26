package proxy

import (
	"fmt"
	"net/rpc"
	"os"

	"openstackcore-rdtagent/lib/resctrl"
)

type ProxyRequest struct {
	Name string
	Res  resctrl.ResAssociation
}

type PipePair struct {
	Reader *os.File
	Writer *os.File
}

func (this *PipePair) Read(p []byte) (int, error) {
	return this.Reader.Read(p)
}

func (this *PipePair) Write(p []byte) (int, error) {
	return this.Writer.Write(p)
}

func (this *PipePair) Close() error {
	this.Writer.Close()
	return this.Reader.Close()
}

type Proxy struct {
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

func RegisterAndServe(pipes PipePair) {
	err := rpc.Register(new(Proxy))
	if err != nil {
		fmt.Println(err)
	}
	rpc.ServeConn(&pipes)
}
