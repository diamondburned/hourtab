package hourtab

import (
	"os"
	"time"

	"go.etcd.io/bbolt"
)

type Session struct {
	db *bbolt.DB
}

type SessionOptions struct {
	DBPath string
	DBMode os.FileMode
	DBOpts *bbolt.Options

	// SyncFrequency controls the frequency that hourtab loops over and saves
	// the state, updating the durations of the projects.
	// TimeoutAfter dictates the number of loops before hourtab marks a project
	// as inactive and stops counting.

	SyncFrequency time.Duration
	TimeoutAfter  uint // SyncFrequency * TimeoutAfter
}

func DefaultOptions() (*SessionOptions, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &SessionOptions{
		DBPath: h,
		DBMode: 0755,

		SyncFrequency: 10 * time.Second,
		TimeoutAfter:  2,
	}, nil
}

func New(opts *SessionOptions) (*Session, error) {
	if opts == nil {
		d, err := DefaultOptions()
		if err != nil {
			return nil, err
		}

		opts = d
	}

	db, err := bbolt.Open(opts.DBPath, opts.DBMode, opts.DBOpts)
	if err != nil {
		return nil, err
	}

	return &Session{
		db: db,
	}, nil
}

func (s *Session) Start() {

}
