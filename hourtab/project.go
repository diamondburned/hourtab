package hourtab

import (
	"sync"
)

type Project struct {
	AbsolutePath string
	GitOrigin    string

	Inactive  bool
	TotalTime uint64 // Nanoseconds
	LastTick  uint64 // UnixNano

	mu sync.Mutex
}

// bump is called when the user writes to anything in this project. timeout is
// the duration in ns, now is the current time in unixnano
func (p *Project) bump(timeout, now uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// if the project was inactive before, aka if the user just started working
	// again
	if p.Inactive {
		// we just assume the user has worked for this long
		p.TotalTime += timeout

		// toggle
		p.Inactive = false

		// set the current tick to now
		p.LastTick = now

		// done
		return
	}

	// if the project isn't inactive before, we just add the time normally

	// first, we calculate the time we need to add
	// then we add the time normally, resetting the lastTick
	p.TotalTime += now - p.LastTick
	p.LastTick = now
}

// this is call independently in the save loop
func (p *Project) timeout(timeout, now uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// if the project is already inactive
	if p.Inactive {
		return // do nothing
	}

	// if it's not
	// check if it needs to be marked inactive

	if p.LastTick+timeout < now {
		// if we're timing out
		p.Inactive = true

		// before quitting, save the progress
		p.TotalTime += timeout

		// We don't set this, as it would mess up the if check
		// p.LastTick = now
	}
}
