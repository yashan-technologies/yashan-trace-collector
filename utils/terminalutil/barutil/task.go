package barutil

import (
	"sync"
)

type task struct {
	sync.Mutex
	name     string
	worker   func() error
	done     chan struct{}
	finished bool
	err      error
}

func (t *task) start() {
	defer func() {
		close(t.done)
	}()
	if t.worker == nil {
		return
	}
	t.err = t.worker()
}

func (t *task) wait() {
	<-t.done
	t.finished = true
}
