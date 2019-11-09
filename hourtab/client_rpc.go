package hourtab

import (
	"net/rpc"

	"gitlab.com/diamondburned/hourtab/ipc"
)

type RPCClient struct {
	*ipc.IPCStatus
	*rpc.Client
}

func NewClient(status *ipc.IPCStatus) (*RPCClient, error) {
	c, err := ipc.NewClient(status)
	if err != nil {
		return nil, err
	}

	return &RPCClient{
		IPCStatus: status,
		Client:    c,
	}, nil
}

func (s *RPCClient) GetAllProjects() (ps []Project, err error) {
	err = s.Call("RPCSession.GetAllProjects", ipc.Nil, &ps)
	return
}

func (s *RPCClient) GetProject(absolutePath string) (p *Project, err error) {
	err = s.Call("RPCSession.GetProject", absolutePath, &p)
	return
}

func (s *RPCClient) AddProject(p *Project) error {
	return s.Call("RPCSession.AddProject", p, &ipc.NilArgument{})
}

func (s *RPCClient) UntrackProject(path string) (deleted bool, err error) {
	err = s.Call("RPCSession.UntrackProject", path, &deleted)
	return
}
