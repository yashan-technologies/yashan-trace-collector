package barutil

import (
	"fmt"
	"io"
	"ytc/defs/bashdef"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type bar struct {
	Name     string
	tasks    []*task
	bar      *mpb.Bar
	progress *Progress
}

type barOption func(b *bar)

func newBar(name string, progress *Progress, opts ...barOption) *bar {
	b := &bar{
		Name:     name,
		tasks:    make([]*task, 0),
		progress: progress,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *bar) addTask(name string, worker func() error) {
	b.tasks = append(b.tasks, &task{
		name:   name,
		worker: worker,
		done:   make(chan struct{}),
	})
}

func (b *bar) draw() {
	efn := func(w io.Writer, s decor.Statistics) (err error) {
		for _, task := range b.tasks {
			if task.finished {
				if task.err == nil {
					if _, err := fmt.Fprintf(w, "\t%s has been %s\n", task.name, bashdef.WithGreen("completed")); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(w, "\t%s has been %s, err: %s\n", task.name, bashdef.WithRed("failed"), task.err.Error()); err != nil {
						return err
					}
				}
			}
		}
		return
	}
	bar := b.progress.mpbProgress.AddBar(int64(len(b.tasks)),
		mpb.BarExtender(mpb.BarFillerFunc(efn), false),
		mpb.PrependDecorators(
			// simple name decorator
			decor.Name(b.Name),
			// decor.DSyncWidth bit enables column width synchronization
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30), "done",
			),
		),
	)
	b.bar = bar
}

func (b *bar) run() {
	defer func() {
		b.progress.wg.Done()
	}()
	for i, t := range b.tasks {
		if i == 0 {
			fmt.Println()
		}
		go t.start()
	}
	for _, t := range b.tasks {
		t.wait()
		b.bar.Increment()
	}
}
