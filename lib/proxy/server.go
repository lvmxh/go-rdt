package proxy

import (
	"fmt"
	"net/rpc"
	"os"
)

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

func RegisterAndServe(pipes PipePair) {
	err := rpc.Register(new(Proxy))
	if err != nil {
		fmt.Println(err)
	}
	rpc.ServeConn(&pipes)
}
