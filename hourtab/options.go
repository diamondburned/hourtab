package hourtab

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type SessionOptions struct {
	DBPath string
	DBMode os.FileMode

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
		DBPath: filepath.Join(h, ".hourtab"),

		SyncFrequency: 10 * time.Second,
		TimeoutAfter:  2,
	}, nil
}

func (opts *SessionOptions) LoadSession() (*Session, error) {
	f, err := os.Open(opts.DBPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open "+opts.DBPath)
	}

	defer f.Close()

	var s Session

	if err := gob.NewDecoder(f).Decode(&s); err != nil {
		return nil, errors.Wrap(err, "Failed to load session from "+opts.DBPath)
	}

	return &s, nil
}
