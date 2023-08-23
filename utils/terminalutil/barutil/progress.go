package barutil

import (
	"fmt"
	"sync"
	"time"

	mpb "github.com/vbauerster/mpb/v8"
)

type ProgressOpt func(p *Progress)

type Progress struct {
	mpbProgress *mpb.Progress
	wg          *sync.WaitGroup
	bars        []*bar
	width       int
}

func WithWidth(width int) ProgressOpt {
	return func(p *Progress) {
		p.width = width
	}
}

func NewProgress(opts ...ProgressOpt) *Progress {
	group := new(sync.WaitGroup)
	p := &Progress{
		wg: group,
	}
	for _, opt := range opts {
		opt(p)
	}
	var mpbOpt []mpb.ContainerOption
	mpbOpt = append(mpbOpt, mpb.WithWaitGroup(group))
	mpbOpt = append(mpbOpt, mpb.WithRefreshRate(time.Microsecond*150))
	if p.width != 0 {
		mpbOpt = append(mpbOpt, mpb.WithWidth(p.width))
	}
	progress := mpb.New(mpbOpt...)
	p.mpbProgress = progress
	return p
}

func (p *Progress) AddBar(name string, namedWorker map[string]func() error) {
	bar := newBar(name, p, withBarWidth(p.width))
	p.wg.Add(1)
	for name, w := range namedWorker {
		bar.addTask(name, w)
	}
	p.bars = append(p.bars, bar)
}

func (p *Progress) Start() {
	for _, bar := range p.bars {
		bar.draw()
		go bar.run()
	}
	p.mpbProgress.Wait()
	// 执行完成下方打印空行
	fmt.Println()
}
