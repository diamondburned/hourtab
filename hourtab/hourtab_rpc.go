package hourtab

import (
	"github.com/pkg/errors"
	"gitlab.com/diamondburned/hourtab/ipc"
)

type RPCSession struct {
	*Session
}

func (s *RPCSession) GetAllProjects(none ipc.NilArgument, reply *[]*Project) error {
	*reply = s.Session.GetAllProjects()
	return nil
}

func (s *RPCSession) GetProject(absolutePath string, reply **Project) error {
	p := s.Session.GetProject(absolutePath)
	if p == nil {
		return errors.New(absolutePath + " not found")
	}

	// Mutex copy does not matter, not like it'll be preserved over IPC anyway.
	*reply = p
	return nil
}

func (s *RPCSession) AddProject(p *Project, none *ipc.NilArgument) error {
	if err := s.Session.AddProject(p); err != nil {
		return err
	}

	*none = ipc.Nil
	return nil
}

func (s *RPCSession) UntrackProject(path string, deleted *bool) error {
	*deleted = s.Session.UntrackProject(path)
	return nil
}
