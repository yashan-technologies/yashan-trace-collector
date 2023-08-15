package barutil

import (
	"fmt"
	"sync"

	"github.com/vbauerster/mpb/v8"
)

type Progress struct {
	mpbProgress *mpb.Progress
	wg          *sync.WaitGroup
	bars        []*bar
}

func NewProgress() *Progress {
	group := new(sync.WaitGroup)
	return &Progress{
		mpbProgress: mpb.New(mpb.WithWaitGroup(group)),
		wg:          group,
	}
}

func (p *Progress) AddBar(name string, namedWorker map[string]func() error) {
	bar := newBar(name, p)
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
