package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/holgerBerger/pfa/pfalib"
)

// LocalProxy is the local endpoint of a proxy to a remote node
type LocalProxy struct {
	node   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// NewLocalProxy creates the local endpoint, and starts the proxy, so it also creates the
// remote end of the proxy
func NewLocalProxy(node string, index int, filename string, blocksize int32, numreaders int, compression pfalib.CompressionType) LocalProxy {
	var err error
	proxy := LocalProxy{node, nil, nil, nil}
	proxy.cmd = exec.Command("/usr/bin/ssh", node, "/bin/cat > /tmp/bla")
	proxy.stdin, err = proxy.cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	proxy.stdout, err = proxy.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	proxy.cmd.Start()
	return proxy
}

func (l LocalProxy) AppendFile(name pfalib.DirEntry) {
	fmt.Println(name.File)
	msg, err := json.Marshal(name)
	if err != nil {
		panic(err)
	}
	l.stdin.Write(msg)
}

func (l LocalProxy) Close() (int64, time.Duration, int64, int64) {
	l.stdin.Close()
	l.stdout.Close()
	err := l.cmd.Wait()
	if err != nil {
		panic(err)
	}
	return 0, 0, 0, 0
}

////////////////////////////////

type RemoteProxy struct {
}

func NewRemoteProxy() RemoteProxy {
	proxy := RemoteProxy{}
	return proxy
}
