package hourtab

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
	"gitlab.com/diamondburned/hourtab/ipc"
)

type Session struct {
	Projects []*Project
	IPC      *ipc.IPCStatus

	opts *SessionOptions

	lock    *flock.Flock
	tick    *time.Ticker
	watcher *watcher.Watcher

	// Projects state RW mutex
	mutex sync.Mutex

	// cancel loop chan
	stop chan struct{}

	// converted timeout from opts
	timeout uint64

	// rpc
	server *rpc.Server
}

func New(opts *SessionOptions) (*Session, error) {
	if opts == nil {
		d, err := DefaultOptions()
		if err != nil {
			return nil, err
		}

		opts = d
	}

	// Acquire the master lock, preventing other hourtab processes from starting
	// up.
	lock := flock.New(opts.DBPath + ".master-lock")

	acquired, err := lock.TryLock()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to acquire master lock")
	}

	if !acquired {
		return nil, errors.New(
			"Master lock already acquired, terminate the other processes!")
	}

	// Make a new IPC server
	server, status, err := ipc.NewServer()
	if err != nil {
		return nil, err
	}

	// Load the old session state
	s, err := opts.LoadSession()
	if err != nil {
		return nil, err
	}

	// Override the old IPC server status
	s.IPC = status

	s.opts = opts
	s.lock = lock
	s.watcher = watcher.New()
	s.stop = make(chan struct{})
	s.timeout = uint64(
		s.opts.SyncFrequency.Nanoseconds() * int64(s.opts.TimeoutAfter))
	s.server = server

	// Finish setting up RPC
	if err := s.server.Register(&RPCSession{s}); err != nil {
		panic("BUG: RPC failed to register: " + err.Error())
	}

	return s, nil
}

/*
	RPC Methods
*/

func (s *Session) GetAllProjects() []*Project {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.Projects
}

func (s *Session) GetProject(absolutePath string) *Project {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, p := range s.Projects {
		if strings.HasPrefix(absolutePath, p.AbsolutePath) {
			return p
		}
	}

	return nil
}

func (s *Session) AddProject(p *Project) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.watcher.AddRecursive(p.AbsolutePath); err != nil {
		return errors.Wrap(err, "Failed to add "+p.GitOrigin)
	}

	s.Projects = append(s.Projects, p)
	return nil
}

// Returns deleted bool
func (s *Session) UntrackProject(path string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, p := range s.Projects {
		if p.AbsolutePath == path {
			// https://github.com/golang/go/wiki/SliceTricks
			s.Projects[i] = s.Projects[len(s.Projects)-1]
			s.Projects[len(s.Projects)-1] = nil
			s.Projects = s.Projects[:len(s.Projects)-1]

			return true
		}
	}

	return false
}

/*
	Normal methods
*/

func (s *Session) Stop() error {
	s.watcher.Close()
	s.stop <- struct{}{}
	return s.lock.Unlock()
}

func (s *Session) Start() {
	go s.startSaveLoop()
	go s.startWatchLoop()
}

func (s *Session) Save() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Acquire the file lock, preventing other clients from reading a
	// being-written state file.
	lock := flock.New(s.opts.DBPath + ".lock")

	if err := lock.Lock(); err != nil {
		return errors.Wrap(err, "Failed to acquire lock")
	}

	defer lock.Unlock()

	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(s); err != nil {
		return errors.Wrap(err, "Failed to save session")
	}

	err := ioutil.WriteFile(s.opts.DBPath, b.Bytes(), s.opts.DBMode)
	if err != nil {
		return errors.Wrap(err, "Failed to write session to "+s.opts.DBPath)
	}

	return nil
}

func (s *Session) startSaveLoop() {
	s.tick = time.NewTicker(s.opts.SyncFrequency)

	var now = time.Now()

	for {
		// check all projects and mark timeout if needed
		s.timeoutProjects(s.timeout, uint64(now.UnixNano()))

		if err := s.Save(); err != nil {
			log.Println("ERROR while saving:", err)
		}

		select {
		case <-s.stop:
			return
		case now = <-s.tick.C:
			continue
		}
	}
}

func (s *Session) startWatchLoop() {
	// add all projects into the watch loop
	for _, p := range s.Projects {
		if err := s.watcher.AddRecursive(p.AbsolutePath); err != nil {
			log.Println("Failed to add", p.GitOrigin+":", err)
		}
	}

	go s.watcher.Start(s.opts.SyncFrequency)

	for {
		select {
		case ev := <-s.watcher.Event:
			if p := s.GetProject(ev.Path); p != nil {
				switch ev.Op {
				case watcher.Create:
					if err := s.watcher.Add(ev.Path); err != nil {
						log.Printf("Failed to add file %s in %s\n",
							ev.Path, p.GitOrigin)
					}
				default:
					now := uint64(time.Now().UnixNano())
					p.bump(s.timeout, now)
				}
			}

		case <-s.watcher.Closed:
			return
		}
	}
}

func (s *Session) timeoutProjects(timeout, now uint64) {
	for _, p := range s.Projects {
		p.timeout(timeout, now)
	}
}
