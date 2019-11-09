package ipc

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
)

var (
	SocketPath = "/tmp/hourtab.sock"
	PortRange  = [2]uint16{49152, 65535}
)

// NilArgument is the placeholder for an empty argument/reply.
type NilArgument struct{}

var Nil = NilArgument{}

type IPCStatus struct {
	SocketPath string
	Port       uint16
}

func (s IPCStatus) IsUnixSocket() bool {
	return s.SocketPath != ""
}

func (s IPCStatus) IsTCP() bool {
	return s.Port > 0
}

func NewServer() (*rpc.Server, *IPCStatus, error) {
	var (
		s   = IPCStatus{}
		l   net.Listener
		err error
	)

	switch runtime.GOOS {
	case "linux", "darwin", "bsd": // Linux, Mac and BSD
		s.SocketPath = SocketPath

		l, err = net.Listen("unix", SocketPath)
		if err != nil {
			err = errors.Wrap(err, "Failed to make a Unix socket")
		}

	default:
		for s.Port = PortRange[0]; s.Port < PortRange[1]; s.Port++ {
			l, err = net.Listen("tcp", ":"+strconv.Itoa(int(s.Port)))
			if err != nil {
				continue
			}
		}

		if err != nil {
			err = fmt.Errorf("Can't find open ports from %d to %d",
				PortRange[0], PortRange[1])
		}
	}

	if err != nil {
		return nil, nil, err
	}

	server := rpc.NewServer()
	server.Accept(l)

	return server, &s, nil
}

func NewClient(s *IPCStatus) (*rpc.Client, error) {
	var (
		rwc io.ReadWriteCloser
		err error
	)

	switch {
	case s.IsUnixSocket():
		rwc, err = net.Dial("unix", s.SocketPath)
	case s.IsTCP():
		rwc, err = net.Dial("tcp", ":"+strconv.Itoa(int(s.Port)))
	}

	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect")
	}

	return rpc.NewClient(rwc), nil
}
