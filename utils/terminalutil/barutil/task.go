package barutil

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"
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
	maxSleepTime := big.NewInt(1000) // 设置最大睡眠时间为1秒
	sleepTime, _ := rand.Int(rand.Reader, maxSleepTime)
	sleepDuration := time.Duration(sleepTime.Int64()) * time.Millisecond
	time.Sleep(sleepDuration)
	t.err = t.worker()
}

func (t *task) wait() {
	<-t.done
	t.finished = true
}
